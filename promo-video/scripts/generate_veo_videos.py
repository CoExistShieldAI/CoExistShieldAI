#!/usr/bin/env python3
# /// script
# requires-python = ">=3.10"
# dependencies = [
#     "google-genai>=2.0.0",
#     "google-auth>=2.0.0",
#     "pillow>=10.0.0",
# ]
# ///
"""
Generate high-fidelity cinematic video clips using Google Cloud Vertex AI Veo 3 (veo-3.1-generate-001).
Reads prompts based on cinematic_prompts_veo_flow.md and outputs MP4s directly to public/cinematics/.
"""

import os
import sys
import time
import shutil
import subprocess
from pathlib import Path
from google import genai
from google.oauth2.credentials import Credentials
from google.genai import types

CINEMATICS_DIR = Path(__file__).parent.parent / "public" / "cinematics"
CINEMATICS_DIR.mkdir(parents=True, exist_ok=True)

PROJECT_ID = "corald"
LOCATION = "us-central1"
MODEL_NAME = "veo-3.1-generate-001"
FALLBACK_MODEL = "veo-3.1-fast-generate-001"

# 7 core cuts defined in cinematic_prompts_veo_flow.md
CLIPS = [
    {
        "id": "cut_1_1",
        "name": "01_tropical_mist.mp4",
        "title": "Deep Tropical Rainforest & Plantation Boundary",
        "prompt": (
            "Cinematic wide establishing drone shot, dense misty tropical rainforest in Southeast Asia at dawn. "
            "The camera moves forward smoothly over lush emerald canopy, revealing a sharp, devastating geometric border "
            "where deep ancient jungle abruptly meets an endless industrial palm oil and sugarcane plantation. "
            "Thick volumetric morning fog, cinematic golden-hour sunlight piercing through towering trees, subtle lens flare, "
            "shot on 35mm Arri Alexa, anamorphic lens, hyperrealistic, dramatic tone"
        ),
        "duration": 8,
    },
    {
        "id": "cut_1_2",
        "name": "02_tiger_stalk.mp4",
        "title": "The Apex Predator's Stalking Gaze",
        "prompt": (
            "Extreme close-up cinematic shot of an endangered wild Sumatran Tiger in low-light dusk, hidden within dense jungle foliage. "
            "Intense, glowing amber eyes staring intently towards the camera, twitching whiskers, wet black nose, hyper-detailed fur texture with raindrops. "
            "The tiger slowly emerges from the shadows, stepping silently onto muddy ground near agricultural fence. "
            "Moody teal and orange rim lighting, shallow depth of field, photorealistic National Geographic documentary style, 8k"
        ),
        "duration": 8,
    },
    {
        "id": "cut_1_3",
        "name": "03_elephant_raid.mp4",
        "title": "Asian Elephants Sugarcane Raid",
        "prompt": (
            "Low-angle cinematic tracking shot at night, massive wild Asian Elephant herd breaking through a wooden plantation fence into a lush sugarcane field. "
            "Flashlight beams cut through the thick dust and darkness, illuminating dust particles and the massive silhouette of the bull elephant with majestic tusks. "
            "High adrenaline, shaky cam documentary feel, cinematic headlights, volumetric dust, intense atmosphere"
        ),
        "duration": 8,
    },
    {
        "id": "cut_1_4",
        "name": "04_roadkill_rain.mp4",
        "title": "Roadkill Carcass & The Looming Spillover",
        "prompt": (
            "Cinematic slow-motion shot, wet dark highway cutting through tropical rainforest during heavy rain. "
            "In the distance, a large animal carcass on the road shoulder, black crows taking flight. "
            "Rain splashing on asphalt, blurred red taillights of an approaching truck reflecting on wet road puddles. "
            "Sinister atmosphere, dark atmospheric grading, foreboding biohazard tension"
        ),
        "duration": 8,
    },
    {
        "id": "cut_4_1",
        "name": "05_japan_bear_persimmon.mp4",
        "aliases": ["05_japan_village.mp4"],
        "title": "Depopulated Japanese Mountain Village & Bear near Persimmon",
        "prompt": (
            "Cinematic wide establishing shot of an abandoned, aging rural Japanese mountain village in late autumn. "
            "Overcast melancholic grey sky, traditional wooden houses with closed shutters, empty roads with fallen leaves. "
            "In the foreground, an overgrown, unharvested persimmon tree heavy with bright glowing orange ripe persimmons (kaki), "
            "a wild Japanese black bear foraging near the tree. "
            "Slow cinematic push-in shot, nostalgic yet eerie solitude, Makoto Shinkai meets realistic cinema tone"
        ),
        "duration": 8,
    },
    {
        "id": "cut_4_2",
        "name": "06_japan_bear_suburb.mp4",
        "aliases": ["06_bear_street.mp4"],
        "title": "Bear Encroaching Residential Suburb",
        "prompt": (
            "Cinematic medium shot at twilight, a massive wild Japanese Black Bear (Tsukinowaguma) with distinct white crescent chest mark, "
            "walking calmly down an empty suburban asphalt street in a Japanese residential neighborhood. "
            "Yellow streetlights reflecting on wet pavement, modern Japanese family houses with parked cars in the background. "
            "Hyperrealistic, tense, news broadcast documentary footage aesthetic"
        ),
        "duration": 8,
    },
    {
        "id": "cut_4_3",
        "name": "bear_prime_minister.mp4",
        "aliases": ["07_bear_pm_inauguration.mp4"],
        "title": "Bear Prime Minister Inauguration",
        "input_image": "bear_prime_minister.jpg",
        "prompt": (
            "Cinematic shot, the dignified female Asian Black Bear Prime Minister standing at the ceremonial mahogany podium "
            "delivers her official speech into the microphones. Her mouth moves naturally as she speaks with solemn gravitas, "
            "nodding slightly and looking confidently out across the press room. "
            "Photographers in the foreground snap photos with blinding camera flashes repeatedly illuminating her dark fur and tailored navy blue business suit. "
            "Realistic movement, authentic press conference speech, camera flashes going off."
        ),
        "duration": 5,
    },
    {
        "id": "cut_3b_1",
        "name": "07_wearos_woman_walking.mp4",
        "title": "Young Worker Walking near Sugarcane & Wrist Vibrate",
        "input_image": "07_wearos_woman_walking.jpg",
        "prompt": (
            "Cinematic medium tracking shot at golden-hour dusk, a young Southeast Asian woman walking along a dusty dirt road adjacent to a towering, dense sugarcane plantation in rural Sumatra. "
            "Atmospheric twilight sunlight piercing through dusty air, gentle breeze swaying the tall green cane stalks. She wears casual light agricultural field workwear. Her right wrist is completely bare with no watch. She wears a modern circular smartwatch (Wear OS) on her left wrist. "
            "The camera tracks smoothly in front of her as she walks. Suddenly, her left wrist smartwatch illuminates with an emergency red alert and vibrates with haptic pulses. "
            "She raises her left wrist to check the alert and halts abruptly with a startled yet focused expression. Photorealistic 8k, shot on 35mm Arri Alexa LF, anamorphic lens, shallow depth of field, dramatic cinematic tone"
        ),
        "duration": 8,
    },
    {
        "id": "cut_3b_2",
        "name": "08_wearos_watch_closeup.mp4",
        "title": "Circular Wear OS Tactical HUD Zoom",
        "input_image": "08_wearos_watch_closeup.jpg",
        "prompt": (
            "Extreme macro close-up cinematic shot of a modern circular Wear OS smartwatch face strapped on a woman's slender wrist. "
            "The bright AMOLED round display is in sharp focus, showing the tactical emergency HUD interface: glowing pulsing red circular radar ring around the bezel edge, high-contrast bold typography displaying '🚨 TIER 3 DANGER // SUMATRAN TIGER 150m AHEAD', and an animated orange proximity arc. "
            "Water droplets and subtle dust particles on the beveled glass screen edge, blurred sugarcane field in the background. Slight hand vibration communicating urgent haptic pulses. "
            "Ultra-sharp 8k macro cinematography, 24fps, high-tech tactical HUD aesthetic"
        ),
        "duration": 7,
    },
    {
        "id": "cut_3b_3",
        "name": "09_woman_leaving_alert.mp4",
        "title": "Vigilant Evacuation & Safe Detour",
        "input_image": "09_woman_leaving_alert.jpg",
        "prompt": (
            "Cinematic tracking shot following the young Southeast Asian woman as she immediately pivots and walks briskly away down the dirt path. "
            "She clutches her watch wrist, glancing tensely over her shoulder into the dark, shadowy depths of the swaying sugarcane field. "
            "Dusk shadows deepen into teal and orange twilight, dust kicking up softly under her boots. Atmospheric suspense thriller cinematography, volumetric dust in evening light, tense motion blur, photorealistic documentary realism"
        ),
        "duration": 8,
    },
    {
        "id": "cut_3b_4",
        "name": "10_tiger_crossing_road.mp4",
        "title": "Peaceful Crossing & Zero Conflict",
        "input_image": "10_tiger_crossing_road.jpg",
        "prompt": (
            "Cinematic wide telephoto shot of the now-completely empty dirt road between the tall sugarcane fields in the deep twilight. "
            "Moments later, a magnificent wild Sumatran Tiger calmly emerges from the dark sugarcane thicket. The tiger walks in slow, graceful, muscular strides directly across the empty dirt road in profile view, rippling orange and black striped fur, long tail swishing gently. "
            "It pauses in the middle of the road, glances toward the distant tree line, and disappears quietly into the opposite jungle reserve. "
            "Soft evening mist rolling across the road, tranquil and majestic atmosphere, zero human conflict, photorealistic National Geographic wildlife documentary 8k"
        ),
        "duration": 8,
    },
]

def get_client():
    token = subprocess.check_output(['gcloud', 'auth', 'print-access-token', '--quiet']).decode().strip()
    creds = Credentials(token)
    return genai.Client(vertexai=True, project=PROJECT_ID, location=LOCATION, credentials=creds)

def generate_clip(client, clip_info, model=MODEL_NAME):
    out_file = CINEMATICS_DIR / clip_info["name"]
    print(f"\n========================================================")
    print(f"🎬 Generating: [{clip_info['id']}] {clip_info['name']}")
    print(f"📌 Title: {clip_info['title']}")
    print(f"🤖 Model: {model}")
    print(f"💬 Prompt: {clip_info['prompt'][:100]}...")
    print(f"========================================================")

    config = types.GenerateVideosConfig(
        aspect_ratio="16:9",
        duration_seconds=clip_info["duration"],
    )

    image_arg = None
    if "input_image" in clip_info:
        img_path = CINEMATICS_DIR / clip_info["input_image"]
        if img_path.exists():
            image_arg = types.Image.from_file(location=str(img_path))
            print(f"🖼️ Using input reference image: {img_path}")

    try:
        op = client.models.generate_videos(
            model=model,
            prompt=clip_info["prompt"],
            image=image_arg,
            config=config,
        )
    except Exception as e:
        print(f"⚠️ Primary model error ({e}). Trying fallback model {FALLBACK_MODEL}...")
        op = client.models.generate_videos(
            model=FALLBACK_MODEL,
            prompt=clip_info["prompt"],
            image=image_arg,
            config=config,
        )

    print(f"🚀 Operation launched: {op.name}")
    print("⏳ Waiting for video generation (polling status)...")

    start_time = time.time()
    while not op.done:
        time.sleep(10)
        try:
            op = client.operations.get(op)
        except Exception as poll_err:
            print(f"  ... transient poll issue: {poll_err}")
            # Refresh client with fresh token in case of expiry
            client = get_client()
            time.sleep(5)
            continue

        elapsed = int(time.time() - start_time)
        print(f"  ... rendering in progress ({elapsed}s elapsed)")

    if op.error:
        print(f"❌ Generation failed with error: {op.error}")
        return False

    if op.response and op.response.generated_videos:
        video_data = op.response.generated_videos[0].video
        video_bytes = video_data.video_bytes
        with open(out_file, "wb") as f:
            f.write(video_bytes)
        file_size_mb = len(video_bytes) / (1024 * 1024)
        print(f"✅ Successfully saved: {out_file} ({file_size_mb:.2f} MB)")

        # Save to aliases if defined
        for alias in clip_info.get("aliases", []):
            alias_path = CINEMATICS_DIR / alias
            shutil.copyfile(out_file, alias_path)
            print(f"   ↳ Created alias: {alias_path}")
        return True
    else:
        print(f"❌ No video returned in response: {op.response}")
        return False

def main():
    print("🎬 Initializing Vertex AI Veo 3 Video Generator...")
    client = get_client()

    target_ids = sys.argv[1:] if len(sys.argv) > 1 else []
    clips_to_run = [c for c in CLIPS if not target_ids or c["id"] in target_ids or c["name"] in target_ids]

    success_count = 0
    total = len(clips_to_run)
    print(f"📋 Running {total} clip(s): {[c['name'] for c in clips_to_run]}")

    for i, clip in enumerate(clips_to_run, 1):
        print(f"\n[Progress: {i}/{total}]")
        success = generate_clip(client, clip)
        if success:
            success_count += 1
        # Short cooldown between generations
        time.sleep(5)

    print(f"\n🎉 Video generation finished: {success_count}/{total} clips generated successfully!")

if __name__ == "__main__":
    main()
