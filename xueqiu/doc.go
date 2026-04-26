// Package xueqiu provides a small client for the Xueqiu realtime quote API.
//
// The client returns the full upstream JSON response as a Quote map so callers
// can read ETF-specific fields such as data.quote.premium_rate without waiting
// for this library to add typed fields.
package xueqiu
