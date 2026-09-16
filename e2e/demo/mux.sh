#!/usr/bin/env bash
set -euo pipefail

# Find latest webm in e2e/recordings/
RAW_VIDEO="$(ls -t e2e/recordings/*.webm | head -n 1)"
AUDIO_TRACK="e2e/demo/audio/master_narration.mp3"

if [ -z "$RAW_VIDEO" ] || [ ! -f "$RAW_VIDEO" ]; then
  echo "Error: No raw webm video found in e2e/recordings/"
  exit 1
fi

echo "=== [MUX] Multiplexing Video and Narration Audio ==="
echo "Video: $RAW_VIDEO"
echo "Audio: $AUDIO_TRACK"

mkdir -p _docs/videos frontend/public

docker run --rm \
  -v "$(pwd)":/workspace \
  -w /workspace \
  python:3.11-alpine \
  sh -c "
    apk add --no-cache ffmpeg > /dev/null 2>&1
    echo '[MUX] 1. Encoding WebM with Opus voice narration...'
    ffmpeg -y -i '$RAW_VIDEO' -i '$AUDIO_TRACK' -c:v copy -c:a libopus -b:a 96k -shortest _docs/videos/choresync_architecture_observability_demo.webm
    
    echo '[MUX] 2. Encoding MP4 with H.264 and AAC voice narration...'
    ffmpeg -y -i '$RAW_VIDEO' -i '$AUDIO_TRACK' -c:v libx264 -preset fast -crf 23 -pix_fmt yuv420p -c:a aac -b:a 128k -shortest -movflags +faststart _docs/videos/choresync_architecture_observability_demo.mp4
    
    cp -f _docs/videos/choresync_architecture_observability_demo.webm frontend/public/choresync_demo.webm
    cp -f _docs/videos/choresync_architecture_observability_demo.mp4 frontend/public/choresync_demo.mp4
    echo '[MUX] Done! Both WebM and MP4 videos with synchronized voice narration are ready.'
  "
