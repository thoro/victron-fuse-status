package main

import (
  "os"
  "fmt"
  "time"
  "strconv"
  "net/http"
  "github.com/prometheus/client_golang/prometheus"
  "github.com/prometheus/client_golang/prometheus/promauto"

  "github.com/prometheus/client_golang/prometheus/promhttp"
  "github.com/thoro/log"
)

var (
  fuse_status = promauto.NewGaugeVec(prometheus.GaugeOpts{
      Name: "fuse_status",
    }, []string{
      "address",
      "fuse",
  })
)

func main () {
  http.Handle("/metrics", promhttp.Handler())

  go func () {
    err := http.ListenAndServe(":2117", nil)

    if err != nil {
      log.Errorf("Unable to serve metrics: %s", err)
      os.Exit(1)
    }
  }()

  ticker := time.NewTicker(10 * time.Second)

  checkFuses(0x08)
  checkFuses(0x09)
  checkFuses(0x0a)
  checkFuses(0x0b)

  for {
    select {
    case <-ticker.C:
      checkFuses(0x08)
      checkFuses(0x09)
      checkFuses(0x0a)
      checkFuses(0x0b)
    }
  }
}

func checkFuses(addr int) {
  s, err := fuseStatus(addr)
  publishFuseStatus(addr, s)

  if err != nil {
    log.Errorf("[%x] Error reading fuse status: %v", addr, err)
  }
}

func publishFuseStatus(addr int, stats []bool) {
  for idx, stat := range stats {
    fuse_status.WithLabelValues(strconv.Itoa(addr), strconv.Itoa(idx)).Set(bool2float(stat))
  }
}

func bool2float(stat bool) float64 {
  if stat {
    return 1
  }

  return 0
}

func fuseStatus(addr int) ([]bool, error) {
  d1, err := NewI2C(uint8(addr), 1)

  if err != nil {
    return nil, err
  }

  defer d1.Close()

  data := []byte{ 0x0 }
  r, err := d1.ReadBytes(data)

  if err != nil {
    return nil, err
  }

  if r != 1 {
    return nil, fmt.Errorf("Not enough bytes")
  }

  return []bool{
    data[0] & 0x10 == 0x10,
    data[0] & 0x20 == 0x20,
    data[0] & 0x40 == 0x40,
    data[0] & 0x80 == 0x80,
  }, nil
}

