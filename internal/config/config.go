package config

import "time"

// Config تنظیمات سرور
type Config struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Option الگوی Functional Options: هر گزینه یک تابع کوچک است
type Option func(*Config)

func WithAddr(addr string) Option {
	return func(c *Config) { c.Addr = addr }
}

func WithTimeouts(read, write, idle time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = read
		c.WriteTimeout = write
		c.IdleTimeout = idle
	}
}

func New(opts ...Option) *Config {
	c := &Config{
		Addr:         ":9988",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}
