#!/data/data/com.termux/files/usr/bin/bash

kill_port() {
  PORT=$1
  if [ -z "$PORT" ]; then
    echo "Usage: kill_port <port>"
    exit 1
  fi

  # Find PID(s) using ss
  PIDS=$(ss -ltnp 2>/dev/null | awk -v port=":$PORT" '$4 ~ port {print $6}' | cut -d',' -f2 | sort -u)

  if [ -z "$PIDS" ]; then
    echo "No process found on port $PORT."
    exit 0
  fi

  echo "Processes using port $PORT:"
  for PID in $PIDS; do
    NAME=$(ps -p $PID -o comm= 2>/dev/null)
    echo "  PID: $PID  Name: $NAME"
  done

  echo
  read -p "Kill ALL these processes? [y/N] " CONFIRM
  case "$CONFIRM" in
  [yY] | [yY][eE][sS])
    for PID in $PIDS; do
      NAME=$(ps -p $PID -o comm= 2>/dev/null)
      kill -9 "$PID"
      echo "Killed PID $PID ($NAME)."
    done
    ;;
  *)
    echo "No processes were killed."
    ;;
  esac
}

kill_port "$@"
