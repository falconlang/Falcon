#!/bin/bash

TARGET="03:01"
echo "⏳ Waiting until $TARGET to run Claude session..."

while true; do
  CURRENT=$(date +"%H:%M")
  if [ "$CURRENT" = "$TARGET" ]; then
    echo "🚀 It's $TARGET! Starting Claude session..."
    claude --resume 84b67571-ee30-49fe-a172-51ee747c5824 \
      --dangerously-skip-permissions \
      --disallowedTools "Bash(rm:*)" "Bash(rmdir:*)" "Bash(sudo:*)" "Bash(chmod:*)" "Bash(dd:*)" \
      --max-turns 80 \
      -p "In the problems/ create the next PROBLEM3.md with next rich unique 500 problem statements, then PROBLEM4.md with next 500 and so on until you are at PROBLEM 20."
    echo "✅ Done!"
    break
  fi
  sleep 30
done
