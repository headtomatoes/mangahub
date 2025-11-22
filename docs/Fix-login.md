# This is command on Windows

re-run docker compose for careful check
run docker exec -i mangahub_db psql -U mangahub -d mangahub < database/health-check.sql
for health check

if any missing notice the run these line respectively:
Get-Content database\migrations\002_fix_user_progress_constraint.up.sql | docker exec -i mangahub_db psql -U mangahub -d mangahub

Get-Content database\migrations\003_add_user_role.up.sql | docker exec -i mangahub_db psql -U mangahub -d mangahub

Get-Content database\migrations\004_add_index_token.up.sql | docker exec -i mangahub_db psql -U mangahub -d mangahub

then the login/register must be proceed with out err
