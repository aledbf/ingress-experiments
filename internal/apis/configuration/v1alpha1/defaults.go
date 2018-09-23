/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	net "net"
	"runtime"

	"github.com/aledbf/ingress-experiments/internal/conversion"
)

const (
	logFormatUpstream = `%v - [$the_real_ip] - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_length $request_time [$proxy_upstream_name] $upstream_addr $upstream_response_length $upstream_response_time $upstream_status $req_id`
)

var (
	gzipTypes = []string{
		"application/atom+xml",
		"application/javascript",
		"application/x-javascript",
		"application/json",
		"application/rss+xml",
		"application/vnd.ms-fontobject",
		"application/x-font-ttf",
		"application/x-web-app-manifest+json",
		"application/xhtml+xml",
		"application/xml",
		"font/opentype image/svg+xml",
		"image/x-icon text/css",
		"text/plain",
		"text/x-component",
	}

	brotliTypes = []string{
		"application/xml+rss",
		"application/atom+xml",
		"application/javascript",
		"application/x-javascript",
		"application/json",
		"application/rss+xml",
		"application/vnd.ms-fontobject",
		"application/x-font-ttf",
		"application/x-web-app-manifest+json",
		"application/xhtml+xml",
		"application/xml",
		"font/opentype",
		"image/svg+xml",
		"image/x-icon",
		"text/css",
		"text/plain",
		"text/x-component",
	}

	// Enabled ciphers list to enabled. The ciphers are specified in the format understood by the OpenSSL library
	// http://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_ciphers
	sslCiphers = []string{
		"ECDHE-ECDSA-AES256-GCM-SHA384",
		"ECDHE-RSA-AES256-GCM-SHA384",
		"ECDHE-ECDSA-CHACHA20-POLY1305",
		"ECDHE-RSA-CHACHA20-POLY1305",
		"ECDHE-ECDSA-AES128-GCM-SHA256",
		"ECDHE-RSA-AES128-GCM-SHA256",
		"ECDHE-ECDSA-AES256-SHA384",
		"ECDHE-RSA-AES256-SHA384",
		"ECDHE-ECDSA-AES128-SHA256",
		"ECDHE-RSA-AES128-SHA256",
	}
)

// NewDefaultConfiguration returns the default nginx configuration
func NewDefaultConfiguration() ConfigurationSpec {

	localhost := &IPAddr{IPAddr: net.IPAddr{IP: net.ParseIP("127.0.0.1")}}
	localhostIPV6 := &IPAddr{IPAddr: net.IPAddr{IP: net.ParseIP("::1")}}

	defNginxStatusIpv4Whitelist := []*IPAddr{localhost}
	defNginxStatusIpv6Whitelist := []*IPAddr{localhostIPV6}

	loadBalancer := LoadBalanceAlgorithm(RoundRobin)

	global := &Global{
		EnableBrotli: conversion.Bool(false),
		BrotliLevel:  conversion.Int(4),
		BrotliTypes:  conversion.StringSlice(brotliTypes),

		EnableGeoIP: conversion.Bool(true),

		EnableGzip: conversion.Bool(true),
		GzipLevel:  conversion.Int(5),
		GzipTypes:  conversion.StringSlice(gzipTypes),

		EnableInfluxDB:             conversion.Bool(false),
		EnableMultiAccept:          conversion.Bool(true),
		EnableProxyProtocol:        conversion.Bool(false),
		EnableRequestID:            conversion.Bool(true),
		EnableReusePort:            conversion.Bool(true),
		EnableUnderscoresInHeaders: conversion.Bool(false),

		IgnoreInvalidHeaders: conversion.Bool(true),

		KeepAlive:         conversion.Int(75),
		KeepAliveRequests: conversion.Int(100),

		LimitConnZoneVariable:  conversion.String("$binary_remote_addr"),
		LimitRequestStatusCode: conversion.Int(429),

		LoadBalanceAlgorithm: &loadBalancer,

		MapHashBucketSize:    conversion.Int(64),
		MaxWorkerConnections: conversion.Int(16384),
		//ProxyProtocolHeaderTimeout: conversion.Time(time.Duration(5) * time.Second),
		RetryNonIdempotent:    conversion.Bool(false),
		ServerNameHashMaxSize: conversion.Int(1024),
		ShowServerTokens:      conversion.Bool(true),
		StatusIPV4Whitelist:   defNginxStatusIpv4Whitelist,
		StatusIPV6Whitelist:   defNginxStatusIpv6Whitelist,

		WorkerProcesses: conversion.Int(runtime.NumCPU()),

		WorkerShutdownTimeout: conversion.Int(10),

		VariablesHashMaxSize: conversion.Int(2048),
		HTTPRedirectCode:     conversion.Int(308),
		NoAuthLocations:      conversion.StringSlice([]string{"/.well-known/acme-challenge"}),
	}

	client := &Client{
		BodyBufferSize:           conversion.String("8k"),
		BodyTimeout:              conversion.Int(60),
		ComputeFullForwardedFor:  conversion.Bool(false),
		ForwardedForHeader:       conversion.String("X-Forwarded-For"),
		HeaderBufferSize:         conversion.String("1k"),
		HeaderTimeout:            conversion.Int(60),
		LargeClientHeaderBuffers: conversion.String("4 8k"),
	}

	http2 := &HTTP2{
		Enabled:       conversion.Bool(true),
		MaxFieldSize:  conversion.String("4k"),
		MaxHeaderSize: conversion.String("16k"),
	}

	/*
		file := &LogFileConfiguration{
			AccessLogPath: conversion.String("/var/log/nginx/access.log"),
			ErrorLogPath:  conversion.String("/var/log/nginx/error.log"),
		}

		syslog := &SyslogConfiguration{
			Enabled: conversion.Bool(false),
			Host:    conversion.String(""),
			Port:    conversion.Int(514),
		}

		log := &Log{
			EnableAccessLog:  conversion.Bool(true),
			ErrorLogLevel:    conversion.String("notice"),
			FormatEscapeJSON: conversion.Bool(false),
			FormatUpstream:   conversion.String(logFormatUpstream),
			File:             file,
			Syslog:           syslog,
		}
	*/

	metrics := &Metrics{
		Enabled: conversion.Bool(false),
		//Latency:        []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		//ResponseLength: []float64{100, 1000, 10000, 100000, 1000000},
		//RequestLength:  []float64{1000, 10000, 100000, 1000000, 10000000},
	}

	snippets := &Snippets{
		Main:     []*Snippet{},
		HTTP:     []*Snippet{},
		Server:   []*Snippet{},
		Location: []*Snippet{},
	}

	hsts := &HSTS{
		Enabled:           conversion.Bool(true),
		IncludeSubdomains: conversion.Bool(true),
		MaxAge:            conversion.Int64(15724800),
		Preload:           conversion.Bool(false),
	}

	ssl := &SSL{
		HSTS: hsts,

		SSLRedirect: conversion.Bool(true),

		Ciphers:   conversion.StringSlice(sslCiphers),
		ECDHCurve: conversion.String("auto"),
		Protocols: conversion.StringSlice([]string{"TLSv1.3", "TLSv1.2"}),

		SessionCache:     conversion.Bool(true),
		SessionCacheSize: conversion.String("10m"),
		SessionTickets:   conversion.Bool(true),
		SessionTimeout:   conversion.String("10m"),

		BufferSize: conversion.String("4k"),
	}

	opentracing := &Opentracing{
		Enabled: conversion.Bool(false),
	}

	upstream := &Upstream{
		AddOriginalURIHeader: conversion.Bool(true),

		BodySize:                      conversion.String("1m"),
		Buffering:                     Off,
		BufferSize:                    conversion.String("4k"),
		CookieDomain:                  conversion.String("off"),
		CookiePath:                    conversion.String("off"),
		ConnectTimeout:                conversion.Int(5),
		EnableServerHeaderFromBackend: conversion.Bool(false),
		FailTimeout:                   conversion.Int(0),
		HeadersHashBucketSize:         conversion.Int(64),
		HeadersHashMaxSize:            conversion.Int(512),
		KeepaliveConnections:          conversion.Int(32),
		MaxFails:                      conversion.Int(0),
		NextUpstream:                  conversion.StringSlice([]string{"error", "timeout"}),
		NextUpstreamTries:             conversion.Int(3),
		ReadTimeout:                   conversion.Int(60),
		RedirectFrom:                  Off,
		RedirectTo:                    Off,
		RequestBuffering:              On,
		SendTimeout:                   conversion.Int(60),
	}

	waf := &WAF{
		EnableModsecurity:    conversion.Bool(false),
		EnableOWASPCoreRules: conversion.Bool(false),
		EnableLuaRestyWAF:    conversion.Bool(true),
	}

	cfg := ConfigurationSpec{
		Global: global,
		Client: client,
		HTTP2:  http2,
		//Log:         log,
		Metrics:     metrics,
		Snippets:    snippets,
		SSL:         ssl,
		Opentracing: opentracing,
		Upstream:    upstream,
		WAF:         waf,
	}

	return cfg
}
