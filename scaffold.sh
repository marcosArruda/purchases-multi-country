#!/usr/bin/env bash



case "$1" in
  "down")
    echo 'Running down!'
    docker compose down
    ;;
  "help")
    echo '$ ./scaffold.sh [down prune build runtests restartapp full-rebuild help]'
    ;;
  "-h")
    echo '$ ./scaffold.sh [down prune build runtests restartapp full-rebuild help]'
    ;;
  "")
    echo '$ ./scaffold.sh [down prune build runtests restartapp full-rebuild help]'
    ;;
  "build")
    echo 'Running build!'
    docker compose build
    ;;
  "restartapp")
    echo 'Running Restart on purchases-multi-country-app!'
    docker compose restart purchases-multi-country
    ;;
  "runtests")
    echo 'Running tests!'
    go test -v -coverpkg=./... -coverprofile=profile.cov ./...; go tool cover -func profile.cov; rm profile.cov
    ;;
  "full-rebuild")
    "$0" down
    "$0" build
    t=1
    shift
    if [[ "$1" != "" ]]; then
      if [[ "$1" == "prune" || "$1" == "-prune" ]]; then
        t=30
        "$0" "$1"
      else
        "$0" "$1"
      fi
    fi
    shift
    if [[ "$1" != "" ]]; then
      if [[ "$1" == "prune" || "$1" == "-prune" ]]; then
        t=30
        "$0" "$1"
      else 
        "$0" "$1"
      fi
    fi
    
    echo 'Running Up'
    docker compose up -d
    echo "waiting $t second(s) til database is ready.."
    sleep $t
    echo 'Running restart purchases-multi-country-app!'
    docker compose restart purchases-multi-country
    echo 'purchases-multi-country-app is Ready!'
    ;;
  "-prune")
    echo 'Running Prune!'
    docker volume prune -f
    ;;
  "prune")
    echo 'Running prune!'
    docker volume prune -f
    ;;
  "-runtests")
    "$0" runtests
    ;;
  "up")
    echo 'Running up!'
    docker compose up -d
    ;;
  "purchase")
    echo 'calling simple /purchases '
    rr=$(echo $RANDOM | md5sum | head -c 10)
    curl -X POST -H 'Content-Type: application/json' -H "Countrycurrency: Brazil-Real" -d "{\"id\": \"$rr\", \"description\": \"Some purchase\", \"amount\": \"20.13\", \"date\": \"2023-10-29\"}" http://localhost:8080/purchases
    echo ""
    ;;
  "get")
    echo 'getting simple /purchases '
    curl -X GET -H 'Content-Type: application/json' -H "Countrycurrency: Brazil-Real" http://localhost:8080/purchases/$2
    echo ""
    ;;
  "getall")
    echo 'getting all /purchases '
    curl -X GET -H 'Content-Type: application/json' -H "Countrycurrency: Brazil-Real" http://localhost:8080/purchases
    echo ""
    ;;
  "logs")
    case "$2" in
      "-app")
        docker logs -f purchases-multi-country
        ;;
      "-db")
        docker logs -f db
        ;;
    esac
    ;;
  *)
    exit 1
    ;;
esac
