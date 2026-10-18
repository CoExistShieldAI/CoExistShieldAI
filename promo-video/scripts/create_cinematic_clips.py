#!/usr/bin/env python3
"""
Convert cinematic images into dynamic camera-motion video clips using ffmpeg.
Creates 30fps MP4 clips with smooth Ken Burns zoom/pan camera movements.
"""

import os
import subprocess
from pathlib import Path

CINEMATICS_DIR = Path(__file__).parent.parent / "public" / "cinematics"

# Definitions for each cinematic cut (duration in seconds, camera motion)
CLIPS = [
    {
        "img": "01_tropical_mist.jpg",
        "out": "01_tropical_mist.mp4",
        "duration": 8.0,
        # Slow drone push-in: zoom from 1.0 to 1.15 centered
        "filter": "zoompan=z='min(zoom+0.0006,1.15)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "02_tiger_stalk.jpg",
        "out": "02_tiger_stalk.mp4",
        "duration": 8.0,
        # Slow zoom towards the tiger's eyes (slightly above center)
        "filter": "zoompan=z='min(zoom+0.0007,1.18)':x='iw/2-(iw/zoom/2)':y='ih*0.4-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "03_elephant_raid.jpg",
        "out": "03_elephant_raid.mp4",
        "duration": 7.5,
        # Tracking pan from left to right with slight zoom
        "filter": "zoompan=z='min(zoom+0.0005,1.12)':x='if(lte(on,1),(iw-iw/zoom)/3,x+0.8)':y='ih/2-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "04_roadkill_rain.jpg",
        "out": "04_roadkill_rain.mp4",
        "duration": 9.0,
        # Slow ominous tilt/push-in
        "filter": "zoompan=z='min(zoom+0.0006,1.15)':x='iw/2-(iw/zoom/2)':y='ih*0.6-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "bear_prime_minister.jpg",
        "out": "bear_prime_minister.mp4",
        "duration": 6.0,
        # Slow push-in on Madam Bear Prime Minister
        "filter": "zoompan=z='min(zoom+0.0006,1.10)':x='iw/2-(iw/zoom/2)':y='ih*0.35-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "05_japan_bear_persimmon.jpg",
        "out": "05_japan_bear_persimmon.mp4",
        "duration": 6.5,
        # Slow zoom in on bear eating persimmon on tree
        "filter": "zoompan=z='min(zoom+0.0007,1.16)':x='iw*0.4-(iw/zoom/2)':y='ih*0.4-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "06_japan_bear_suburb.jpg",
        "out": "06_japan_bear_suburb.mp4",
        "duration": 6.5,
        # Tracking push-in along the suburban road
        "filter": "zoompan=z='min(zoom+0.0006,1.14)':x='iw/2-(iw/zoom/2)':y='ih*0.55-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "07_wearos_woman_walking.jpg",
        "out": "07_wearos_woman_walking.mp4",
        "duration": 6.0,
        # Tracking shot of woman walking at dusk
        "filter": "zoompan=z='min(zoom+0.0006,1.12)':x='iw/2-(iw/zoom/2)':y='ih*0.4-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "08_wearos_watch_closeup.jpg",
        "out": "08_wearos_watch_closeup.mp4",
        "duration": 6.0,
        # Slow macro zoom into watch dial
        "filter": "zoompan=z='min(zoom+0.0008,1.18)':x='iw/2-(iw/zoom/2)':y='ih/2-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "09_woman_leaving_alert.jpg",
        "out": "09_woman_leaving_alert.mp4",
        "duration": 6.0,
        # Tracking shot of woman retreating cautiously
        "filter": "zoompan=z='min(zoom+0.0005,1.12)':x='iw*0.45-(iw/zoom/2)':y='ih*0.4-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    },
    {
        "img": "10_tiger_crossing_road.jpg",
        "out": "10_tiger_crossing_road.mp4",
        "duration": 7.0,
        # Slow cinematic pan as tiger crosses empty road
        "filter": "zoompan=z='min(zoom+0.0005,1.10)':x='if(lte(on,1),(iw-iw/zoom)/3,x+0.9)':y='ih/2-(ih/zoom/2)':d={frames}:s=1920x1080:fps=30"
    }
]

def main():
    for clip in CLIPS:
        img_path = CINEMATICS_DIR / clip["img"]
        out_path = CINEMATICS_DIR / clip["out"]
        
        if not img_path.exists():
            print(f"Skipping {clip['img']} (not found yet)")
            continue

        frames = int(clip["duration"] * 30)
        vf = clip["filter"].format(frames=frames)
        
        print(f"🎬 Creating dynamic cinematic clip: {clip['out']} ({clip['duration']}s)...")
        cmd = [
            "ffmpeg", "-y",
            "-loop", "1",
            "-i", str(img_path),
            "-vf", vf,
            "-t", str(clip["duration"]),
            "-c:v", "libx264",
            "-pix_fmt", "yuv420p",
            str(out_path)
        ]
        res = subprocess.run(cmd, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
        if res.returncode == 0:
            print(f"  ✓ Saved to {out_path}")
        else:
            print(f"  ✗ Error: {res.stderr.decode()[:200]}")

if __name__ == "__main__":
    main()
