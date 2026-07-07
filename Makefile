title ?=
page ?=
genre ?=
sort ?=
page_size ?=

get-all:
	@curl -s -G "localhost:4000/v1/movies" \
		$(if $(title),--data-urlencode "title=$(title)") \
		$(if $(genre),--data-urlencode "genres=$(genre)") \
		$(if $(page),--data-urlencode "page=$(page)") \
		$(if $(page_size),--data-urlencode "page_size=$(page_size)") \
		$(if $(sort),--data-urlencode "sort=$(sort)") | jq

