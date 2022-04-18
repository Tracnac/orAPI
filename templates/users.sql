select *
from dba_users
where 1=1
{{ range .username }} and username = {{ . }} {{ end }}
