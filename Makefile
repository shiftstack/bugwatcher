
query_url_%:
	@sed '1,/SHIFTSTACK_QUERY/d;/^)$$/,$$d;s/\s\+"\(.*\)"/\1/' $*.py | sed ':a; N; $$!ba; s/\n//g'
