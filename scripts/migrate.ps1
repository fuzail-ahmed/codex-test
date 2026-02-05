param(
  [string]$Dsn,
  [string]$Command = "up"
)

if (-not $Dsn) {
  Write-Error "-Dsn is required"
  exit 1
}

go run ./cmd/migrate -database $Dsn -command $Command