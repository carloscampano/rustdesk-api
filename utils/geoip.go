package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type GeoInfo struct {
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	ISP         string `json:"isp"`
	Flag        string `json:"flag"`
	Private     bool   `json:"private,omitempty"`
	Error       bool   `json:"error,omitempty"`
}

const geoCacheTTL = 24 * time.Hour
const geoMaxBatch = 40

type geoCacheEntry struct {
	info GeoInfo
	exp  time.Time
}

var (
	geoMu    sync.Mutex
	geoCache = map[string]geoCacheEntry{}
	geoHTTP  = &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
)

func LookupGeoBatch(ips []string) map[string]GeoInfo {
	out := make(map[string]GeoInfo, len(ips))
	if len(ips) > geoMaxBatch {
		ips = ips[:geoMaxBatch]
	}
	seen := map[string]struct{}{}
	for _, raw := range ips {
		ip := ParseIP(raw)
		if ip == nil {
			continue
		}
		key := ip.String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out[key] = lookupOne(ip)
	}
	return out
}

func lookupOne(ip net.IP) GeoInfo {
	s := ip.String()
	if !IsPublicIP(s) {
		return GeoInfo{Country: "Red local", ISP: "Privada", Private: true}
	}
	if g, ok := geoFromCache(s); ok {
		return g
	}
	g, err := fetchIPWho(s)
	if err != nil {
		return GeoInfo{Error: true}
	}
	geoToCache(s, g)
	return g
}

func geoFromCache(ip string) (GeoInfo, bool) {
	geoMu.Lock()
	defer geoMu.Unlock()
	e, ok := geoCache[ip]
	if !ok || time.Now().After(e.exp) {
		return GeoInfo{}, false
	}
	return e.info, true
}

func geoToCache(ip string, g GeoInfo) {
	geoMu.Lock()
	defer geoMu.Unlock()
	if len(geoCache) > 5000 {
		geoCache = map[string]geoCacheEntry{}
	}
	geoCache[ip] = geoCacheEntry{info: g, exp: time.Now().Add(geoCacheTTL)}
}

func fetchIPWho(ip string) (GeoInfo, error) {
	u := fmt.Sprintf("https://ipwho.is/%s?lang=es&fields=success,country,country_code,connection,flag", ip)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return GeoInfo{}, err
	}
	req.Header.Set("User-Agent", "rustdesk-api-geo")
	resp, err := geoHTTP.Do(req)
	if err != nil {
		return GeoInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return GeoInfo{}, fmt.Errorf("ipwho status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return GeoInfo{}, err
	}
	var raw struct {
		Success    bool   `json:"success"`
		Country    string `json:"country"`
		CountryCode string `json:"country_code"`
		Connection struct {
			ISP string `json:"isp"`
			Org string `json:"org"`
		} `json:"connection"`
		Flag struct {
			Emoji string `json:"emoji"`
		} `json:"flag"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return GeoInfo{}, err
	}
	if !raw.Success {
		return GeoInfo{}, fmt.Errorf("ipwho unsuccessful")
	}
	isp := raw.Connection.ISP
	if isp == "" {
		isp = raw.Connection.Org
	}
	return GeoInfo{
		Country:     raw.Country,
		CountryCode: raw.CountryCode,
		ISP:         isp,
		Flag:        raw.Flag.Emoji,
	}, nil
}
