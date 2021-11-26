
pretriage_query_url:
	@sed '1,/SHIFTSTACK_QUERY/d;/^)$$/,$$d;s/\s\+"\(.*\)"/\1/' pretriage.py  | sed ':a; N; $$!ba; s/\n//g'
.PHONY: pretriage_query_url

posttriage_query_url:
	@sed '1,/SHIFTSTACK_QUERY/d;/^)$$/,$$d;s/\s\+"\(.*\)"/\1/' posttriage.py | sed ':a; N; $$!ba; s/\n//g'
.PHONY: posttriage_query_url
