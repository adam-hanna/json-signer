module github.com/tbruyelle/legacykey

go 1.22.2

require (
	github.com/99designs/keyring v1.2.1
	github.com/bgentry/speakeasy v0.1.0
)

replace github.com/99designs/keyring => /home/tom/gohack/github.com/99designs/keyring // was github.com/99designs/keyring => github.com/cosmos/keyring v1.2.0

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dvsekhvalnov/jose2go v1.6.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/term v0.19.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
