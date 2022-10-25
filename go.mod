module github.com/Nigh/picture-to-array

go 1.18

require (
	github.com/google/uuid v1.3.0
	github.com/hotei/bmp v0.0.0-20150430041436-f620cebab0c7
	github.com/rubenfonseca/fastimage v0.0.0-20170112075114-7e006a27a95b
	localhost/picarray v0.0.0-00010101000000-000000000000
)

replace localhost/picarray => ./picarray
