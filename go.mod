module github.com/ollybritton/prox

go 1.13

require (
	github.com/Bogdan-D/go-socks4 v0.0.0-20160129084303-092515145880
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/fatih/color v1.7.0
	github.com/frankban/quicktest v1.6.0 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/logrusorgru/aurora v0.0.0-20191017060258-dc85c304c434
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/olekukonko/tablewriter v0.0.2
	github.com/ollybritton/prox/providers v0.0.0-00010101000000-000000000000
	github.com/pierrec/lz4 v2.3.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/stretchr/testify v1.4.0
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/net v0.0.0-20191109021931-daa7c04131f5
)

replace github.com/ollybritton/prox/providers => ./providers
