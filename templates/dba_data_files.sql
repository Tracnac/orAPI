select *
from dba_data_files
where 1=1
{{ range .tablespace_name }} and tablespace_name = {{.}} {{ end }}