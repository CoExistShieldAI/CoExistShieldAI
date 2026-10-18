#!/usr/bin/env python3
"""
Generate high-fidelity CoExist Shield AI brand icons, favicons, PWA assets,
and native mobile icons (iOS / Android / Wear OS).
"""

import os
import subprocess
from pathlib import Path
from PIL import Image

BASE_DIR = Path(__file__).resolve().parent.parent

# 1. Master SVG Definitions
SVG_FAVICON_TRANSPARENT = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" width="512" height="512">
  <defs>
    <linearGradient id="shieldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#10B981" />
      <stop offset="50%" stop-color="#06B6D4" />
      <stop offset="100%" stop-color="#3B82F6" />
    </linearGradient>
    <linearGradient id="shieldFill" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#0F172A" stop-opacity="0.95" />
      <stop offset="100%" stop-color="#020617" stop-opacity="0.98" />
    </linearGradient>
    <linearGradient id="checkGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#34D399" />
      <stop offset="100%" stop-color="#38BDF8" />
    </linearGradient>
    <filter id="neonGlow" x="-20%" y="-20%" width="140%" height="140%">
      <feGaussianBlur stdDeviation="12" result="blur" />
      <feComposite in="SourceGraphic" in2="blur" operator="over" />
    </filter>
    <filter id="coreGlow" x="-30%" y="-30%" width="160%" height="160%">
      <feGaussianBlur stdDeviation="8" result="blur" />
      <feComposite in="SourceGraphic" in2="blur" operator="over" />
    </filter>
  </defs>

  <!-- Outer Neon Glow Path -->
  <path d="M256 42 L396 98 C396 240 336 376 256 462 C176 376 116 240 116 98 Z"
        fill="none" stroke="url(#shieldGrad)" stroke-width="28" stroke-linejoin="round"
        filter="url(#neonGlow)" opacity="0.8" />

  <!-- Shield Body -->
  <path d="M256 46 L392 100 C392 238 334 372 256 456 C178 372 120 238 120 100 Z"
        fill="url(#shieldFill)" stroke="url(#shieldGrad)" stroke-width="16" stroke-linejoin="round" />

  <!-- Inner Radar / Tactical Concentric Arcs -->
  <path d="M176 210 A 100 100 0 0 1 336 210" fill="none" stroke="#06B6D4" stroke-width="4" stroke-dasharray="6 8" opacity="0.4" />
  <path d="M196 230 A 70 70 0 0 1 316 230" fill="none" stroke="#10B981" stroke-width="3" stroke-dasharray="4 6" opacity="0.5" />

  <!-- Central AI CoExist Check & Nodes -->
  <!-- Upper AI Pulse Core Node -->
  <circle cx="256" cy="154" r="10" fill="#38BDF8" filter="url(#coreGlow)" />
  <circle cx="256" cy="154" r="5" fill="#FFFFFF" />

  <!-- Connection line to checkmark apex -->
  <line x1="256" y1="164" x2="256" y2="216" stroke="#06B6D4" stroke-width="3" stroke-dasharray="3 3" opacity="0.6" />

  <!-- Tactical Safety Checkmark (Protection + Verification) -->
  <path d="M184 266 L236 320 L332 208"
        fill="none" stroke="url(#checkGrad)" stroke-width="26" stroke-linecap="round" stroke-linejoin="round"
        filter="url(#coreGlow)" />

  <!-- Satellite Sensor Nodes on Shield Boundary -->
  <circle cx="256" cy="46" r="6" fill="#34D399" />
  <circle cx="392" cy="100" r="5" fill="#06B6D4" />
  <circle cx="120" cy="100" r="5" fill="#10B981" />
  <circle cx="256" cy="456" r="5" fill="#3B82F6" />
</svg>
"""

SVG_APP_ICON_OPAQUE = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1024 1024" width="1024" height="1024">
  <defs>
    <!-- Background Gradients -->
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#020617" />
      <stop offset="50%" stop-color="#0B132B" />
      <stop offset="100%" stop-color="#020617" />
    </linearGradient>
    <radialGradient id="centerGlow" cx="50%" cy="45%" r="60%">
      <stop offset="0%" stop-color="#064E3B" stop-opacity="0.55" />
      <stop offset="50%" stop-color="#083344" stop-opacity="0.25" />
      <stop offset="100%" stop-color="#020617" stop-opacity="0" />
    </radialGradient>

    <!-- Shield & Element Gradients -->
    <linearGradient id="shieldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#10B981" />
      <stop offset="40%" stop-color="#06B6D4" />
      <stop offset="80%" stop-color="#3B82F6" />
      <stop offset="100%" stop-color="#6366F1" />
    </linearGradient>
    <linearGradient id="shieldFill" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#0F172A" stop-opacity="0.95" />
      <stop offset="100%" stop-color="#020617" stop-opacity="0.98" />
    </linearGradient>
    <linearGradient id="checkGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#34D399" />
      <stop offset="50%" stop-color="#38BDF8" />
      <stop offset="100%" stop-color="#60A5FA" />
    </linearGradient>

    <!-- Filters -->
    <filter id="neonGlowLarge" x="-30%" y="-30%" width="160%" height="160%">
      <feGaussianBlur stdDeviation="32" result="blur1" />
      <feGaussianBlur stdDeviation="16" result="blur2" />
      <feMerge>
        <feMergeNode in="blur1" />
        <feMergeNode in="blur2" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>
    <filter id="coreGlowLarge" x="-40%" y="-40%" width="180%" height="180%">
      <feGaussianBlur stdDeviation="24" result="blur1" />
      <feGaussianBlur stdDeviation="10" result="blur2" />
      <feMerge>
        <feMergeNode in="blur1" />
        <feMergeNode in="blur2" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>

    <!-- Subtle Tactical Grid Pattern -->
    <pattern id="grid" width="48" height="48" patternUnits="userSpaceOnUse">
      <path d="M 48 0 L 0 0 0 48" fill="none" stroke="#1E293B" stroke-width="1" stroke-opacity="0.4" />
    </pattern>
  </defs>

  <!-- Base Solid Background (No Alpha for App Stores) -->
  <rect width="1024" height="1024" fill="url(#bgGrad)" />
  <rect width="1024" height="1024" fill="url(#centerGlow)" />
  <rect width="1024" height="1024" fill="url(#grid)" />

  <!-- Tactical Decorative Rings -->
  <circle cx="512" cy="512" r="440" fill="none" stroke="#334155" stroke-width="1.5" stroke-dasharray="8 16" opacity="0.3" />
  <circle cx="512" cy="512" r="390" fill="none" stroke="#06B6D4" stroke-width="2" stroke-dasharray="12 24" opacity="0.25" />

  <!-- Outer Neon Halo for Shield -->
  <path d="M512 110 L792 220 C792 504 672 776 512 948 C352 776 232 504 232 220 Z"
        fill="none" stroke="url(#shieldGrad)" stroke-width="56" stroke-linejoin="round"
        filter="url(#neonGlowLarge)" opacity="0.75" />

  <!-- Shield Body Surface -->
  <path d="M512 118 L784 224 C784 500 668 768 512 936 C356 768 240 500 240 224 Z"
        fill="url(#shieldFill)" stroke="url(#shieldGrad)" stroke-width="32" stroke-linejoin="round" />

  <!-- Inner Concentric Tactical Radar Arcs -->
  <path d="M352 440 A 200 200 0 0 1 672 440" fill="none" stroke="#06B6D4" stroke-width="8" stroke-dasharray="14 18" opacity="0.4" />
  <path d="M392 480 A 140 140 0 0 1 632 480" fill="none" stroke="#10B981" stroke-width="6" stroke-dasharray="10 14" opacity="0.5" />
  <path d="M432 520 A 90 90 0 0 1 592 520" fill="none" stroke="#38BDF8" stroke-width="4" stroke-dasharray="6 10" opacity="0.6" />

  <!-- Upper AI Pulse Core Node & Network Lines -->
  <circle cx="512" cy="326" r="24" fill="#38BDF8" filter="url(#coreGlowLarge)" />
  <circle cx="512" cy="326" r="12" fill="#FFFFFF" />

  <!-- Sensor connection line -->
  <line x1="512" y1="350" x2="512" y2="450" stroke="#06B6D4" stroke-width="6" stroke-dasharray="6 6" opacity="0.7" />

  <!-- Tactical Safety Checkmark (Protection + Verification) -->
  <path d="M368 550 L472 656 L664 432"
        fill="none" stroke="url(#checkGrad)" stroke-width="52" stroke-linecap="round" stroke-linejoin="round"
        filter="url(#coreGlowLarge)" />

  <!-- Boundary Tactical Nodes -->
  <circle cx="512" cy="118" r="14" fill="#34D399" filter="url(#coreGlowLarge)" />
  <circle cx="784" cy="224" r="12" fill="#06B6D4" />
  <circle cx="240" cy="224" r="12" fill="#10B981" />
  <circle cx="512" cy="936" r="12" fill="#3B82F6" />

  <!-- High-Tech Corner Accent Brackets -->
  <path d="M120 180 L120 120 L180 120" fill="none" stroke="#06B6D4" stroke-width="6" stroke-linecap="round" opacity="0.5" />
  <path d="M904 180 L904 120 L844 120" fill="none" stroke="#06B6D4" stroke-width="6" stroke-linecap="round" opacity="0.5" />
  <path d="M120 844 L120 904 L180 904" fill="none" stroke="#10B981" stroke-width="6" stroke-linecap="round" opacity="0.5" />
  <path d="M904 844 L904 904 L844 904" fill="none" stroke="#10B981" stroke-width="6" stroke-linecap="round" opacity="0.5" />
</svg>
"""

# Android Adaptive Icon Foreground (432x432, centered icon within 288x288 safe zone)
SVG_ANDROID_FOREGROUND = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 432 432" width="432" height="432">
  <defs>
    <linearGradient id="shieldGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#10B981" />
      <stop offset="50%" stop-color="#06B6D4" />
      <stop offset="100%" stop-color="#3B82F6" />
    </linearGradient>
    <linearGradient id="shieldFill" x1="0%" y1="0%" x2="0%" y2="100%">
      <stop offset="0%" stop-color="#0F172A" stop-opacity="0.95" />
      <stop offset="100%" stop-color="#020617" stop-opacity="0.98" />
    </linearGradient>
    <linearGradient id="checkGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#34D399" />
      <stop offset="100%" stop-color="#38BDF8" />
    </linearGradient>
    <filter id="coreGlow" x="-30%" y="-30%" width="160%" height="160%">
      <feGaussianBlur stdDeviation="6" result="blur" />
      <feComposite in="SourceGraphic" in2="blur" operator="over" />
    </filter>
  </defs>

  <!-- Scaled down to fit within ~260x260 in center (216, 216) -->
  <g transform="translate(216, 216) scale(0.62) translate(-256, -256)">
    <!-- Shield Outline & Body -->
    <path d="M256 46 L392 100 C392 238 334 372 256 456 C178 372 120 238 120 100 Z"
          fill="url(#shieldFill)" stroke="url(#shieldGrad)" stroke-width="20" stroke-linejoin="round" />

    <!-- Tactical Radar Arcs -->
    <path d="M176 210 A 100 100 0 0 1 336 210" fill="none" stroke="#06B6D4" stroke-width="6" stroke-dasharray="8 10" opacity="0.5" />

    <!-- Upper AI Pulse Core Node -->
    <circle cx="256" cy="154" r="14" fill="#38BDF8" filter="url(#coreGlow)" />
    <circle cx="256" cy="154" r="7" fill="#FFFFFF" />

    <!-- Tactical Safety Checkmark -->
    <path d="M184 266 L236 320 L332 208"
          fill="none" stroke="url(#checkGrad)" stroke-width="32" stroke-linecap="round" stroke-linejoin="round"
          filter="url(#coreGlow)" />

    <!-- Nodes -->
    <circle cx="256" cy="46" r="8" fill="#34D399" />
    <circle cx="392" cy="100" r="7" fill="#06B6D4" />
    <circle cx="120" cy="100" r="7" fill="#10B981" />
    <circle cx="256" cy="456" r="7" fill="#3B82F6" />
  </g>
</svg>
"""

# Android Adaptive Icon Background (432x432)
SVG_ANDROID_BACKGROUND = """<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 432 432" width="432" height="432">
  <defs>
    <linearGradient id="bgGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" stop-color="#020617" />
      <stop offset="50%" stop-color="#0B132B" />
      <stop offset="100%" stop-color="#020617" />
    </linearGradient>
    <radialGradient id="centerGlow" cx="50%" cy="50%" r="55%">
      <stop offset="0%" stop-color="#064E3B" stop-opacity="0.6" />
      <stop offset="60%" stop-color="#083344" stop-opacity="0.25" />
      <stop offset="100%" stop-color="#020617" stop-opacity="0" />
    </radialGradient>
  </defs>
  <rect width="432" height="432" fill="url(#bgGrad)" />
  <rect width="432" height="432" fill="url(#centerGlow)" />
  <circle cx="216" cy="216" r="160" fill="none" stroke="#1E293B" stroke-width="2" stroke-dasharray="6 12" opacity="0.4" />
</svg>
"""


def render_svg_to_png(svg_string: str, output_png_path: Path, width: int, height: int):
    """Render an SVG string to a PNG file using rsvg-convert."""
    temp_svg = output_png_path.with_suffix(".tmp.svg")
    temp_svg.write_text(svg_string, encoding="utf-8")
    try:
        subprocess.run(
            ["rsvg-convert", "-w", str(width), "-h", str(height), "-f", "png", "-o", str(output_png_path), str(temp_svg)],
            check=True,
            capture_output=True
        )
    finally:
        if temp_svg.exists():
            temp_svg.unlink()


def generate_ico(png_16: Path, png_32: Path, png_48: Path, output_ico: Path):
    """Generate multi-resolution .ico from 16, 32, 48 PNGs using Pillow."""
    img16 = Image.open(png_16)
    img32 = Image.open(png_32)
    img48 = Image.open(png_48)
    img48.save(output_ico, format="ICO", sizes=[(16, 16), (32, 32), (48, 48)], append_images=[img16, img32])


def main():
    print("=== Generating CoExist Shield AI Brand Icons ===")
    out_dir = BASE_DIR / "docs" / "branding"
    out_dir.mkdir(parents=True, exist_ok=True)

    # Save Master SVGs
    (out_dir / "favicon.svg").write_text(SVG_FAVICON_TRANSPARENT, encoding="utf-8")
    (out_dir / "app_icon_master.svg").write_text(SVG_APP_ICON_OPAQUE, encoding="utf-8")
    (out_dir / "android_foreground.svg").write_text(SVG_ANDROID_FOREGROUND, encoding="utf-8")
    (out_dir / "android_background.svg").write_text(SVG_ANDROID_BACKGROUND, encoding="utf-8")
    print("✓ Saved Master SVGs in docs/branding/")

    # Generate Intermediate PNGs
    tmp_dir = BASE_DIR / "scripts" / "branding_assets"
    tmp_dir.mkdir(parents=True, exist_ok=True)

    render_svg_to_png(SVG_FAVICON_TRANSPARENT, tmp_dir / "fav-16.png", 16, 16)
    render_svg_to_png(SVG_FAVICON_TRANSPARENT, tmp_dir / "fav-32.png", 32, 32)
    render_svg_to_png(SVG_FAVICON_TRANSPARENT, tmp_dir / "fav-48.png", 48, 48)

    # Full app icon renders (with rich tactical background)
    render_svg_to_png(SVG_APP_ICON_OPAQUE, tmp_dir / "icon-180.png", 180, 180) # apple-touch-icon
    render_svg_to_png(SVG_APP_ICON_OPAQUE, tmp_dir / "icon-192.png", 192, 192) # PWA
    render_svg_to_png(SVG_APP_ICON_OPAQUE, tmp_dir / "icon-512.png", 512, 512) # PWA
    render_svg_to_png(SVG_APP_ICON_OPAQUE, tmp_dir / "icon-1024.png", 1024, 1024) # iOS App Store

    # Create multi-size .ico
    generate_ico(tmp_dir / "fav-16.png", tmp_dir / "fav-32.png", tmp_dir / "fav-48.png", tmp_dir / "favicon.ico")
    print("✓ Generated multi-size favicon.ico (16, 32, 48)")

    # 1. Deploy to web-admin
    web_admin_public = BASE_DIR / "web-admin" / "public"
    web_admin_public.mkdir(parents=True, exist_ok=True)
    (web_admin_public / "favicon.svg").write_text(SVG_FAVICON_TRANSPARENT, encoding="utf-8")
    for f in ["favicon.ico", "icon-180.png", "icon-192.png", "icon-512.png"]:
        data = (tmp_dir / f).read_bytes()
        dest_name = "apple-touch-icon.png" if f == "icon-180.png" else f
        (web_admin_public / dest_name).write_bytes(data)
    print("✓ Deployed icons to web-admin/public/")

    # 2. Deploy to ionic-app
    ionic_public = BASE_DIR / "ionic-app" / "public"
    ionic_public.mkdir(parents=True, exist_ok=True)
    (ionic_public / "favicon.svg").write_text(SVG_FAVICON_TRANSPARENT, encoding="utf-8")
    for f in ["favicon.ico", "icon-180.png", "icon-192.png", "icon-512.png"]:
        data = (tmp_dir / f).read_bytes()
        dest_name = "apple-touch-icon.png" if f == "icon-180.png" else f
        (ionic_public / dest_name).write_bytes(data)

    # iOS AppIcon (1024x1024)
    ios_icon_dest = BASE_DIR / "ionic-app" / "ios" / "App" / "App" / "Assets.xcassets" / "AppIcon.appiconset" / "AppIcon-512@2x.png"
    if ios_icon_dest.parent.exists():
        ios_icon_dest.write_bytes((tmp_dir / "icon-1024.png").read_bytes())
        print("✓ Updated iOS AppIcon-512@2x.png (1024x1024)")

    # 3. Deploy Android Icons (composeApp & wearApp)
    android_scales = {
        "mipmap-mdpi": 48,
        "mipmap-hdpi": 72,
        "mipmap-xhdpi": 96,
        "mipmap-xxhdpi": 144,
        "mipmap-xxxhdpi": 192,
    }

    # Render adaptive icon foreground / background
    render_svg_to_png(SVG_ANDROID_FOREGROUND, tmp_dir / "ic_launcher_foreground.png", 432, 432)
    render_svg_to_png(SVG_ANDROID_BACKGROUND, tmp_dir / "ic_launcher_background.png", 432, 432)

    adaptive_xml = """<?xml version="1.0" encoding="utf-8"?>
<adaptive-icon xmlns:android="http://schemas.android.com/apk/res/android">
    <background android:drawable="@drawable/ic_launcher_background" />
    <foreground android:drawable="@drawable/ic_launcher_foreground" />
</adaptive-icon>
"""

    for target in ["composeApp", "wearApp"]:
        res_dir = BASE_DIR / target / "src" / "androidMain" / "res"
        res_dir.mkdir(parents=True, exist_ok=True)

        drawable_dir = res_dir / "drawable"
        drawable_dir.mkdir(parents=True, exist_ok=True)
        (drawable_dir / "ic_launcher_foreground.png").write_bytes((tmp_dir / "ic_launcher_foreground.png").read_bytes())
        (drawable_dir / "ic_launcher_background.png").write_bytes((tmp_dir / "ic_launcher_background.png").read_bytes())

        # Adaptive icon xml
        anydpi_dir = res_dir / "mipmap-anydpi-v26"
        anydpi_dir.mkdir(parents=True, exist_ok=True)
        (anydpi_dir / "ic_launcher.xml").write_text(adaptive_xml, encoding="utf-8")
        (anydpi_dir / "ic_launcher_round.xml").write_text(adaptive_xml, encoding="utf-8")

        # Legacy mipmaps
        for folder, sz in android_scales.items():
            mdir = res_dir / folder
            mdir.mkdir(parents=True, exist_ok=True)
            render_svg_to_png(SVG_APP_ICON_OPAQUE, mdir / "ic_launcher.png", sz, sz)
            render_svg_to_png(SVG_APP_ICON_OPAQUE, mdir / "ic_launcher_round.png", sz, sz)

        print(f"✓ Deployed Android mipmap & adaptive icons to {target}")

    # Cleanup temporary directory
    if tmp_dir.exists():
        import shutil
        shutil.rmtree(tmp_dir)

    print("\nAll brand icons generated and deployed successfully!")


if __name__ == "__main__":
    main()
