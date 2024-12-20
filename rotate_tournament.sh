YESTERDAY=$(date -u -d "yesterday" +"%Y-%m-%d")

# End yesterday's tournament
curl -X PUT "https://good-blast-real.fly.dev/tournaments/end/${YESTERDAY}"

# Start today's tournament
curl -X POST "https://good-blast-real.fly.dev/tournaments/start"
