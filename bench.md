\hotspot-tracker> go test -bench .
goos: windows
goarch: amd64
pkg: github.com/aayush993/hotspot-tracker    
cpu: Intel(R) Core(TM) i7-4510U CPU @ 2.00GHz
BenchmarkRecordRequest-4                  167506              6514 ns/op               0 B/op          0 allocs/op
BenchmarkGetHotspots-4                   4319992               303.4 ns/op           416 B/op          1 allocs/op
BenchmarkIsHotspot-4                    23821812                50.38 ns/op            0 B/op          0 allocs/op
BenchmarkConcurrentAccess-4                 6958            167575 ns/op          181647 B/op        189 allocs/op
BenchmarkEnhancedConcurrentAccess-4        10000            144679 ns/op            8638 B/op         53 allocs/op
PASS
ok      github.com/aayush993/hotspot-tracker    12.383s