package graphite

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/msaf1980/go-metrics"
	"github.com/msaf1980/go-stringutils"
)

// Config provides a container with configuration parameters for
// the Graphite exporter
type Config struct {
	Host           string        `toml:"host" yaml:"host" json:"host"`                                  // Network address to connect to
	FlushInterval  time.Duration `toml:"interval" yaml:"interval" json:"interval"`                      // Flush interval
	DurationUnit   time.Duration `toml:"duration" yaml:"duration" json:"duration"`                      // Time conversion unit for durations
	Prefix         string        `toml:"prefix" yaml:"prefix" json:"prefix"`                            // Prefix to be prepended to metric names
	TagPrefix      string        `toml:"tag_prefix" yaml:"tag_prefix" json:"tag_prefix"`                // Prefix to be prepended to metric name tag
	ConnectTimeout time.Duration `toml:"connect_timeout" yaml:"connect_timeout" json:"connect_timeout"` // Connect timeout
	Timeout        time.Duration `toml:"timeout" yaml:"timeout" json:"timeout"`                         // Write timeout
	Retry          int           `toml:"retry" yaml:"retry" json:"retry"`                               // Reconnect retry count
	BufSize        int           `toml:"buffer" yaml:"buffer" json:"buffer"`                            // Buffer size

	MinLock bool `toml:"min_lock" yaml:"min_lock" json:"min_lock"` // Minimize time of read-locking of metric registry (but with some costs), set if application do dynamic metrics register/unregister

	Percentiles []float64 `toml:"percentiles" yaml:"percentiles" json:"percentiles"` // Percentiles to export from timers and histograms
	percentiles []string  `toml:"-" yaml:"-" json:"-"`                               // Percentiles keys (pregenerated)

	Tags map[string]string `toml:"tags" yaml:"tags" json:"tags"` // Tags for sended metrics (not used directky in graphite client, merge it with  metric individual tags)
}

func setDefaults(c *Config) {
	if c.ConnectTimeout == 0 {
		c.ConnectTimeout = time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = time.Second
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = time.Minute
	}
	if c.DurationUnit <= 0 {
		c.DurationUnit = time.Millisecond
	}
	if c.BufSize <= 0 {
		c.BufSize = 1024
	}
	if c.Retry <= 0 {
		c.Retry = 1
	}
	c.percentiles = make([]string, 0, len(c.Percentiles))
	for _, p := range c.Percentiles {
		key := strings.Replace(strconv.FormatFloat(p*100.0, 'f', -1, 64), ".", "", 1)
		c.percentiles = append(c.percentiles, "."+key+"-percentile")
	}
}

func loggerSucces() {
	log.Printf("graphite: success")
}

func loggerError(err error) {
	log.Printf("graphite: %v", err)
}

type Graphite struct {
	c    *Config
	conn net.Conn
	buf  stringutils.Builder

	loggerSuccess func()
	loggerError   func(error)

	stop chan struct{}
	wg   sync.WaitGroup
}

// Graphite is a blocking exporter function which reports metrics in r
// to a graphite server located at addr, flushing them every d duration
// and prepending metric names with prefix.
func New(flushInterval time.Duration, prefix string, host string, timeout time.Duration) *Graphite {
	return WithConfig(&Config{
		Host:           host,
		FlushInterval:  flushInterval,
		DurationUnit:   time.Nanosecond,
		Prefix:         prefix,
		Percentiles:    []float64{0.5, 0.75, 0.95, 0.99, 0.999},
		Timeout:        timeout,
		ConnectTimeout: timeout,
	})
}

func newGraphite(c *Config) *Graphite {
	setDefaults(c)
	return &Graphite{
		c:             c,
		loggerSuccess: loggerSucces,
		loggerError:   loggerError,
	}
}

// WithConfig is a blocking exporter function just like Graphite,
// but it takes a GraphiteConfig instead.
func WithConfig(c *Config) *Graphite {
	g := newGraphite(c)
	g.buf.Grow(c.BufSize)
	return g
}

// Once performs a single submission to Graphite, returning a
// non-nil error on failed connections. This can be used in a loop
// similar to GraphiteWithConfig for custom error handling.
func Once(c *Config, r metrics.Registry) error {
	g := newGraphite(c)
	g.buf.Grow(c.BufSize)
	err := g.send(r)
	g.Close()
	return err
}

func (g *Graphite) SetLoggerSucces(f func()) {
	g.loggerSuccess = f
}

func (g *Graphite) SetLoggerError(f func(error)) {
	g.loggerError = f
}

func (g *Graphite) Start(r metrics.Registry) {
	g.wg.Add(1)
	g.stop = make(chan struct{})
	go func() {
		var lastErr bool
		var err error
		defer g.wg.Done()
		t := time.NewTicker(g.c.FlushInterval)
	LOOP:
		for {
			select {
			case <-t.C:
				if err = g.send(r); err != nil {
					log.Println(err)
				}
			case <-g.stop:
				break LOOP
			}
		}
		if err = g.Close(); err == nil {
			if lastErr {
				lastErr = false
				g.loggerSuccess()
			}
		} else if !lastErr {
			lastErr = true
			g.loggerError(err)
		}
	}()
}

func (g *Graphite) Stop() {
	g.stop <- struct{}{}
	g.wg.Wait()
}

func (g *Graphite) Close() error {
	err := g.flush()
	if g.conn != nil {
		g.conn.Close()
	}
	return err
}

func (g *Graphite) connect() error {
	var err error
	for i := 0; i < g.c.Retry; i++ {
		g.conn, err = net.DialTimeout("tcp", g.c.Host, g.c.ConnectTimeout)
		if nil == err {
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return err
}

func (g *Graphite) writeIntMetric(name, postfix string, tags string, v, ts int64) (err error) {
	if tags == "" {
		if g.c.Prefix != "" {
			g.buf.WriteString(g.c.Prefix)
			g.buf.WriteRune('.')
		}
	} else if g.c.TagPrefix != "" {
		g.buf.WriteString(g.c.TagPrefix)
		g.buf.WriteRune('.')
	}
	g.buf.WriteString(name)
	g.buf.WriteString(postfix)
	g.buf.WriteString(tags)
	g.buf.WriteRune(' ')
	g.buf.WriteInt(v, 10)
	g.buf.WriteRune(' ')
	g.buf.WriteInt(ts, 10)
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) writeUintMetric(name, postfix string, tags string, v uint64, ts int64) (err error) {
	if tags == "" {
		if g.c.Prefix != "" {
			g.buf.WriteString(g.c.Prefix)
			g.buf.WriteRune('.')
		}
	} else if g.c.TagPrefix != "" {
		g.buf.WriteString(g.c.TagPrefix)
		g.buf.WriteRune('.')
	}
	g.buf.WriteString(name)
	g.buf.WriteString(postfix)
	g.buf.WriteString(tags)
	g.buf.WriteRune(' ')
	g.buf.WriteUint(v, 10)
	g.buf.WriteRune(' ')
	g.buf.WriteInt(ts, 10)
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) writeHistogramMetric(name, label, le string, tags string, v uint64, ts int64) (err error) {
	if tags == "" {
		if g.c.Prefix != "" {
			g.buf.WriteString(g.c.Prefix)
			g.buf.WriteRune('.')
		}
		g.buf.WriteString(name)
		g.buf.WriteString(label)
	} else {
		if g.c.TagPrefix != "" {
			g.buf.WriteString(g.c.TagPrefix)
			g.buf.WriteRune('.')
		}
		g.buf.WriteString(name)
		g.buf.WriteString(label)
		g.buf.WriteString(tags)
		if le != "" {
			g.buf.WriteString(";le=")
			g.buf.WriteString(le)
		}
	}
	g.buf.WriteRune(' ')
	g.buf.WriteUint(v, 10)
	g.buf.WriteRune(' ')
	g.buf.WriteInt(ts, 10)
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) writeFloatMetric(name, postfix string, tags string, v float64, ts int64) (err error) {
	if tags == "" {
		if g.c.Prefix != "" {
			g.buf.WriteString(g.c.Prefix)
			g.buf.WriteRune('.')
		}
	} else if g.c.TagPrefix != "" {
		g.buf.WriteString(g.c.TagPrefix)
		g.buf.WriteRune('.')
	}
	g.buf.WriteString(name)
	g.buf.WriteString(postfix)
	g.buf.WriteString(tags)
	g.buf.WriteRune(' ')
	g.buf.WriteFloat(v, 'f', 2, 64)
	g.buf.WriteRune(' ')
	g.buf.WriteInt(ts, 10)
	g.buf.WriteRune('\n')

	if g.c.BufSize <= g.buf.Len() {
		return g.flush()
	}
	return nil
}

func (g *Graphite) flush() (err error) {
	if g.buf.Len() > 0 {
		if g.conn == nil {
			if err = g.connect(); err != nil {
				return
			}
		}
		g.conn.SetWriteDeadline(time.Now().Add(g.c.Timeout))
		_, err = g.conn.Write(g.buf.Bytes())
		if err != nil {
			if err = g.connect(); err != nil {
				return
			}
			_, err = g.conn.Write(g.buf.Bytes())
		}
		if err == nil {
			g.buf.Reset()
		}
	}

	return
}

func (g *Graphite) send(r metrics.Registry) error {
	if nil == r {
		r = metrics.DefaultRegistry
	}

	var err error

	now := time.Now().Unix()
	// du := float64(g.c.DurationUnit)
	// flushSeconds := float64(g.c.FlushInterval) / float64(time.Second)

	if g.conn == nil {
		if err = g.connect(); err != nil {
			return err
		}
	}

	if err = g.flush(); err != nil {
		return err
	}

	r.Each(func(name, tags string, tagsMap map[string]string, i interface{}) error {
		switch metric := i.(type) {
		case metrics.Counter:
			// count := metric.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			if err = g.writeUintMetric(name, "", tags, metric.Count(), now); err != nil {
				return err
			}
		case metrics.DownCounter:
			// count := metric.Count()
			// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
			if err = g.writeIntMetric(name, "", tags, metric.Count(), now); err != nil {
				return err
			}
		case metrics.Gauge:
			// fmt.Fprintf(w, "%s.%s.value %d %d\n", c.Prefix, name, metric.Value(), now)
			if err = g.writeIntMetric(name, "", tags, metric.Value(), now); err != nil {
				return err
			}
		case metrics.UGauge:
			if err = g.writeUintMetric(name, "", tags, metric.Value(), now); err != nil {
				return err
			}
		case metrics.FGauge:
			// fmt.Fprintf(w, "%s.%s.value %f %d\n", c.Prefix, name, metric.Value(), now)
			if err = g.writeFloatMetric(name, "", tags, metric.Value(), now); err != nil {
				return err
			}
		case metrics.Healthcheck:
			if err = g.writeIntMetric(name, "", tags, int64(metric.Check()), now); err != nil {
				return err
			}
		case metrics.HistogramInterface:
			vals := metric.Values()
			leAliases := metric.WeightsAliases()
			if metric.IsSummed() {
				for i, label := range metric.Labels() {
					if err = g.writeHistogramMetric(name, label, leAliases[i], tags, vals[i], now); err != nil {
						return err
					}
				}
				if err = g.writeHistogramMetric(name, metric.NameTotal(), "", tags, vals[0], now); err != nil {
					return err
				}
			} else {
				var total uint64
				for i, label := range metric.Labels() {
					if err = g.writeHistogramMetric(name, label, leAliases[i], tags, vals[i], now); err != nil {
						return err
					}
					total += vals[i]
				}
				if err = g.writeHistogramMetric(name, metric.NameTotal(), "", tags, total, now); err != nil {
					return err
				}
			}
		case metrics.Rate:
			v, rate := metric.Values()
			if err = g.writeIntMetric(name, metric.Name(), tags, v, now); err != nil {
				return err
			}
			if err = g.writeFloatMetric(name, metric.RateName(), tags, rate, now); err != nil {
				return err
			}
		case metrics.FRate:
			v, rate := metric.Values()
			if err = g.writeFloatMetric(name, metric.Name(), tags, v, now); err != nil {
				return err
			}
			if err = g.writeFloatMetric(name, metric.RateName(), tags, rate, now); err != nil {
				return err
			}
		// case metrics.Histogram:
		// 	h := metric.Snapshot()
		// 	ps := h.Percentiles(g.c.Percentiles)
		// 	// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, h.Count(), now)
		// 	g.writeIntMetric(name, ".count", tags, h.Count(), now)
		// 	// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, h.Min(), now)
		// 	g.writeIntMetric(name, ".min", tags, h.Min(), now)
		// 	// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, h.Max(), now)
		// 	g.writeIntMetric(name, ".max", tags, h.Max(), now)
		// 	// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, h.Mean(), now)
		// 	g.writeFloatMetric(name, ".mean", tags, h.Mean(), now)
		// 	// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, h.StdDev(), now)
		// 	g.writeFloatMetric(name, ".std-dev", tags, h.StdDev(), now)
		// 	for psIdx, psKey := range g.c.percentiles {
		// 		// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		// 		// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx], now)
		// 		g.writeFloatMetric(name, psKey, tags, ps[psIdx], now)
		// 	}
		// case metrics.Meter:
		// 	m := metric.Snapshot()
		// 	// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, m.Count(), now)
		// 	g.writeIntMetric(name, ".count", tags, m.Count(), now)
		// 	// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, m.Rate1(), now)
		// 	g.writeFloatMetric(name, ".one-minute", tags, m.Rate1(), now)
		// 	// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, m.Rate5(), now)
		// 	g.writeFloatMetric(name, ".five-minute", tags, m.Rate5(), now)
		// 	// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, m.Rate15(), now)
		// 	g.writeFloatMetric(name, ".fifteen-minute", tags, m.Rate15(), now)
		// 	// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, m.RateMean(), now)
		// 	g.writeFloatMetric(name, ".mean", tags, m.RateMean(), now)
		// case metrics.Timer:
		// 	t := metric.Snapshot()
		// 	ps := t.Percentiles(g.c.Percentiles)
		// 	count := t.Count()
		// 	// fmt.Fprintf(w, "%s.%s.count %d %d\n", c.Prefix, name, count, now)
		// 	g.writeIntMetric(name, ".count", tags, count, now)
		// 	// fmt.Fprintf(w, "%s.%s.count_ps %.2f %d\n", c.Prefix, name, float64(count)/flushSeconds, now)
		// 	g.writeFloatMetric(name, ".count_ps", tags, float64(count)/flushSeconds, now)
		// 	// fmt.Fprintf(w, "%s.%s.min %d %d\n", c.Prefix, name, t.Min()/int64(du), now)
		// 	g.writeIntMetric(name, ".min", tags, t.Min()/int64(du), now)
		// 	// fmt.Fprintf(w, "%s.%s.max %d %d\n", c.Prefix, name, t.Max()/int64(du), now)
		// 	g.writeIntMetric(name, ".max", tags, t.Max()/int64(du), now)
		// 	// fmt.Fprintf(w, "%s.%s.mean %.2f %d\n", c.Prefix, name, t.Mean()/du, now)
		// 	g.writeFloatMetric(name, ".mean", tags, t.Mean()/du, now)
		// 	// fmt.Fprintf(w, "%s.%s.std-dev %.2f %d\n", c.Prefix, name, t.StdDev()/du, now)
		// 	g.writeFloatMetric(name, ".std-dev", tags, t.StdDev()/du, now)
		// 	for psIdx, psKey := range g.c.percentiles {
		// 		// key := strings.Replace(strconv.FormatFloat(psKey*100.0, 'f', -1, 64), ".", "", 1)
		// 		// fmt.Fprintf(w, "%s.%s.%s-percentile %.2f %d\n", c.Prefix, name, key, ps[psIdx]/du, now)
		// 		g.writeFloatMetric(name, psKey, tags, ps[psIdx]/du, now)
		// 	}
		// 	// fmt.Fprintf(w, "%s.%s.one-minute %.2f %d\n", c.Prefix, name, t.Rate1(), now)
		// 	g.writeFloatMetric(name, ".one-minute", tags, t.Rate1(), now)
		// 	// fmt.Fprintf(w, "%s.%s.five-minute %.2f %d\n", c.Prefix, name, t.Rate5(), now)
		// 	g.writeFloatMetric(name, ".five-minute", tags, t.Rate5(), now)
		// 	// fmt.Fprintf(w, "%s.%s.fifteen-minute %.2f %d\n", c.Prefix, name, t.Rate15(), now)
		// 	g.writeFloatMetric(name, ".fifteen-minute", tags, t.Rate15(), now)
		// 	// fmt.Fprintf(w, "%s.%s.mean-rate %.2f %d\n", c.Prefix, name, t.RateMean(), now)
		// 	g.writeFloatMetric(name, ".mean-rate", tags, t.RateMean(), now)
		default:
			g.loggerError(fmt.Errorf("unable to record metric of type %T", i))
		}
		return nil
	}, g.c.MinLock)
	return g.flush()
}
