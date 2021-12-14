module mosaicmfg.com/stl-to-3mf

go 1.16

require mosaicmfg.com/stl-to-3mf/ps3mf v0.0.0

replace (
	mosaicmfg.com/stl-to-3mf/ps3mf => ./ps3mf
	mosaicmfg.com/stl-to-3mf/util => ./util
)
