
query_url:
	@sed '1,/SHIFTSTACK_QUERY/d;/^)$$/,$$d;s/\s\+"\(.*\)"/\1/' pretriage/main.py | sed ':a; N; $$!ba; s/\n//g'
.PHONY: query_url
