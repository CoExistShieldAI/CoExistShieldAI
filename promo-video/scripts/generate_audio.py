#!/usr/bin/env python3
"""
Audio & Script Generation Script for CoExist Shield AI Promo Video
Generates TTS speech audio clips using edge-tts with Singapore & US neural voices.
"""

import asyncio
import json
import os
import subprocess
from pathlib import Path

VOICE_NARRATOR = "en-US-ChristopherNeural"  # Deep, authoritative cinematic trailer voice
VOICE_Ranger_Tora = "en-SG-WayneNeural"    # Energetic, friendly Singaporean English game-streamer voice

OUTPUT_DIR = Path(__file__).parent.parent / "public" / "audio"
OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

# Full Script: 3-4 minutes timed segments
SCRIPT = [
    # --- PART 1: CINEMATIC HOOK & THE APAC CRISIS ---
    {
        "id": "p1_01_mist",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Along the forest frontiers of Southeast Asia, an invisible conflict escalates between human expansion and wildlife."
    },
    {
        "id": "p1_02_tiger",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Displaced apex predators, driven by territory loss, stalk the boundaries of villages and commercial plantations."
    },
    {
        "id": "p1_03_elephant",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Massive elephant herds, lured by ripe sugar crops, face frightened farmers in devastating night collisions."
    },
    {
        "id": "p1_04_roadkill",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "And unmanaged roadside carcasses become deadly predatory magnets, threatening communities with fatal encounters and zoonotic spillover."
    },

    # --- PART 2: INTRODUCING COEXIST SHIELD AI & GEOFENCING ---
    {
        "id": "p2_01_intro",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Hey APAC innovators! Ranger Tora here with CoExist Shield AI—our autonomous platform for human-wildlife coexistence!"
    },
    {
        "id": "p2_02_geofence",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "First, what is geofencing? Think of it as an invisible virtual fence drawn on a map using GPS. Enter the zone, and your phone or watch triggers an alert! But traditional geofences are fixed and static—completely useless against moving wildlife! That is why CoExist Shield AI introduces Dynamic Predictive Geofencing: an autonomous AI safety shield that moves with animals, predicting paths and alerting communities before encounters ever happen!"
    },
    {
        "id": "p2_03_algorithm",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Look under the hood at our Geospatial Inference Brain! Ingesting crop phenology, carcass decay, wind scent plumes, and thermal hotspots, our neural risk engine synthesizes these priors into dynamic probability surfaces, projecting Tier 1 to Tier 3 geofences in real time!"
    },

    # --- PART 3: LIVE SYSTEM DEMO (PLAYWRIGHT CAPTURE) ---
    {
        "id": "p3_01_dashboard",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Check out our live command dashboard! Here in southern Thailand, our causal model tracks crop harvest phenology."
    },
    {
        "id": "p3_02_elephant_map",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "See those glowing orange clusters? That's our Crop Attractant Map pinpointing ripening sugarcane and jackfruit. The AI forecasts elephant foraging corridors days in advance!"
    },
    {
        "id": "p3_03_tiger_corral",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Next up, Sumatra, Indonesia. Tigers don't wander randomly; they follow livestock grazing corrals and fresh roadkill carcasses."
    },
    {
        "id": "p3_04_geofence_alert",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Boom! A Tier 3 Danger Geofence triggers automatically around a fresh kill site! Field workers receive instant tactile alerts on their phones and smartwatches, completely avoiding ambush zones!"
    },
    {
        "id": "p3_05_wearos_haptic",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Walking near a sugarcane plantation at dusk, a young worker suddenly feels an intense haptic pulse on her wrist. Checking her Wear OS watch, the tactical HUD flashes red: Tier 3 Danger, Tiger 150 meters ahead!"
    },
    {
        "id": "p3_06_wearos_retreat",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Cautious and alert, she immediately retreats down a safe detour route. Moments later, a wild tiger quietly crosses the empty road. Thanks to CoExist Shield AI, not a single person was harmed, and the tiger passed safely into the wild."
    },

    # --- PART 4: JAPAN EXPANSION & THE BEAR PRIME MINISTER CINEMATIC ---
    {
        "id": "p4_01_japan_crisis",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "Our algorithm scales seamlessly across ecosystems. In depopulating rural Japan, human retreat and abandoned fruit trees have led to unprecedented bear encroachments into residential valleys."
    },
    {
        "id": "p4_02_bear_pm_joke",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "At this rate... Japan's next Prime Minister might just be a bear!"
    },
    {
        "id": "p4_03_tora_reaction",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Whoa, hold on! As sharp as Madam Prime Minister looks, that's NOT our roadmap! Haha!"
    },
    {
        "id": "p4_04_japan_dx",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "Jokes aside, our AI analyzes bear incident records to pinpoint the top five percent highest-risk abandoned persimmon trees, providing municipal teams with precision removal targets funded via Government Crowdfunding!"
    },

    # --- PART 5: ONE HEALTH IMPACT & CONCLUSION ---
    {
        "id": "p5_01_onehealth",
        "speaker": "Ranger Tora",
        "voice": VOICE_Ranger_Tora,
        "text": "By clearing roadkill and stopping illegal bushmeat trade before it starts, CoExist Shield AI acts as an autonomous frontline shield against future zoonotic pandemics."
    },
    {
        "id": "p5_02_grand_finale",
        "speaker": "Narrator",
        "voice": VOICE_NARRATOR,
        "text": "From tropical Asian jungles to temperate Japanese valleys, CoExist Shield AI protects biodiversity, safeguards human lives, and powers harmonious coexistence. Thank you, Singapore!"
    }
]

async def generate_clip(item):
    out_file = OUTPUT_DIR / f"{item['id']}.mp3"
    print(f"Generating [{item['speaker']}]: {item['id']} ...")
    cmd = [
        "uvx", "edge-tts",
        "--voice", item["voice"],
        "--text", item["text"],
        "--write-media", str(out_file)
    ]
    proc = await asyncio.create_subprocess_exec(*cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    await proc.communicate()
    
    # Calculate duration using ffprobe
    ffprobe_cmd = [
        "ffprobe", "-v", "error", "-show_entries", "format=duration",
        "-of", "default=noprint_wrappers=1:nokey=1", str(out_file)
    ]
    duration_proc = await asyncio.create_subprocess_exec(*ffprobe_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, _ = await duration_proc.communicate()
    duration = float(out.decode().strip() or "0")
    
    return {
        "id": item["id"],
        "speaker": item["speaker"],
        "voice": item["voice"],
        "text": item["text"],
        "file": f"/audio/{item['id']}.mp3",
        "duration": duration
    }

async def main():
    tasks = [generate_clip(item) for item in SCRIPT]
    results = await asyncio.gather(*tasks)
    
    total_audio_time = sum(r["duration"] for r in results)
    print(f"\nAll TTS generated successfully! Total spoken duration: {total_audio_time:.1f} seconds ({total_audio_time/60:.2f} mins)")
    
    meta_path = OUTPUT_DIR / "timeline.json"
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(results, f, indent=2, ensure_ascii=False)
    print(f"Timeline metadata saved to {meta_path}")

if __name__ == "__main__":
    asyncio.run(main())
