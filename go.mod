module github.com/ollybritton/prox

go 1.13

require (
	github.com/Bogdan-D/go-socks4 v0.0.0-20160129084303-092515145880
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/fatih/color v1.7.0
	github.com/logrusorgru/aurora v0.0.0-20191017060258-dc85c304c434
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/nsf/termbox-go v0.0.0-20190817171036-93860e161317 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/ollybritton/prox/providers v0.0.0-00010101000000-000000000000
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/uber-common/cpustat v0.0.0-20191203071534-b3265a2cd987 // indirect
	github.com/uber-common/termui v0.0.0-20160224010800-5eb574feb7a3 // indirect
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5
)

replace github.com/ollybritton/prox/providers => ./providers
