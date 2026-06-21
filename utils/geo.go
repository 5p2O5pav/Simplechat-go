package utils

import (
    "net"
    "github.com/oschwald/geoip2-golang"
    "log"
)

var db *geoip2.Reader

func InitGeoIP(dbPath string) error {
    var err error
    db, err = geoip2.Open(dbPath)
    if err != nil {
        return err
    }
    return nil
}

func CloseGeoIP() {
    if db != nil {
        db.Close()
    }
}

func GetGeoInfo(ipStr string) string {
    if db == nil {
        return "未知"
    }
    ip := net.ParseIP(ipStr)
    if ip == nil {
        return "未知"
    }
    record, err := db.City(ip)
    if err != nil {
        return "未知"
    }
    country := record.Country.Names["zh-CN"]
    if country == "" {
        country = record.Country.Names["en"]
    }
    city := record.City.Names["zh-CN"]
    if city == "" {
        city = record.City.Names["en"]
    }
    if country == "" && city == "" {
        return "未知"
    }
    if country == "" {
        return city
    }
    if city == "" {
        return country
    }
    return country + " " + city
}
