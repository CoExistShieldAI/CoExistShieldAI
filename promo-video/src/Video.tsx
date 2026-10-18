import React from 'react';
import {
  Audio,
  Img,
  interpolate,
  OffthreadVideo,
  Sequence,
  staticFile,
  useCurrentFrame,
} from 'remotion';
import { WipeAvatar } from './components/WipeAvatar';
import { Subtitles } from './components/Subtitles';
import { WearOSWatchModal } from './components/WearOSWatchModal';
import { AlgorithmSimulation } from './components/AlgorithmSimulation';
import { GeofenceExplainer } from './components/GeofenceExplainer';

interface TimelineItem {
  id: string;
  speaker: string;
  voice: string;
  text: string;
  file: string;
  duration: number;
}

interface PromoVideoProps {
  timeline: TimelineItem[];
}

const FPS = 30;

export const PromoVideo: React.FC<PromoVideoProps> = ({ timeline }) => {
  let accumulatedFrames = 0;

  return (
    <div
      style={{
        width: 1920,
        height: 1080,
        backgroundColor: '#050811',
        overflow: 'hidden',
        position: 'relative',
        fontFamily: 'Inter, system-ui, sans-serif',
      }}
    >
      {timeline.map((item, index) => {
        const durationInFrames = Math.ceil(item.duration * FPS);
        const fromFrame = accumulatedFrames;
        accumulatedFrames += durationInFrames;

        return (
          <Sequence
            key={item.id}
            from={fromFrame}
            durationInFrames={durationInFrames}
            name={`${item.id} - ${item.speaker}`}
          >
            {/* Audio Voice Track */}
            <Audio src={staticFile(item.file)} volume={1.0} />

            {/* Scene Visuals based on ID */}
            <SceneContent item={item} durationInFrames={durationInFrames} />

            {/* Subtitles Overlay */}
            <Subtitles speaker={item.speaker} text={item.text} />

            {/* Ranger Tora Wipe Avatar (Shown during Tora's dialogue or Part 2/3/4) */}
            {(item.speaker === 'Ranger Tora' || item.id.startsWith('p3_')) && (
              <WipeAvatar isSpeaking={item.speaker === 'Ranger Tora'} />
            )}
          </Sequence>
        );
      })}
    </div>
  );
};

// Scene-specific visual rendering component
const SceneContent: React.FC<{ item: TimelineItem; durationInFrames: number }> = ({ item, durationInFrames }) => {
  const frame = useCurrentFrame();

  // Part 3.5 WearOS Live Demonstration: Sugarcane Worker Alert & Tiger Crossing Road
  if (item.id === 'p3_05_wearos_haptic') {
    const cutFrame = 120; // Cut to watch closeup after 4.0s

    return (
      <div style={{ width: '100%', height: '100%', position: 'relative', overflow: 'hidden', backgroundColor: '#000' }}>
        {/* Shot 1: Woman walking near sugarcane at dusk */}
        <Sequence from={0} durationInFrames={cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/07_wearos_woman_walking.mp4')}
            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
          />
        </Sequence>

        {/* Shot 2: Macro watch closeup (starts from frame 0 of video, in slow motion) */}
        <Sequence from={cutFrame} durationInFrames={durationInFrames - cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/08_wearos_watch_closeup.mp4')}
            playbackRate={0.5}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
              filter: 'blur(4px) brightness(0.6)',
            }}
          />
        </Sequence>

        {/* Anamorphic letterbox bars */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 60, backgroundColor: '#000', zIndex: 40 }} />
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 60, backgroundColor: '#000', zIndex: 40 }} />

        {/* Location Tag */}
        <div style={{ position: 'absolute', top: 80, left: 80, zIndex: 45 }}>
          <div style={{ color: '#EF4444', fontSize: 16, fontWeight: 800, letterSpacing: '3px' }}>
            ● WEAR OS LIVE HAPTIC TRIGGER
          </div>
          <div style={{ color: '#fff', fontSize: 13, fontFamily: 'monospace' }}>
            📍 RIAU, SUMATRA // SUGARCANE & PALM BORDER
          </div>
        </div>

        {/* Big Smartwatch HUD Pop-in on Screen */}
        <Sequence from={cutFrame} durationInFrames={durationInFrames - cutFrame} layout="none">
          <WearOSWatchModal distance="150m" threat="SUMATRAN TIGER" />
        </Sequence>
      </div>
    );
  }

  if (item.id === 'p3_06_wearos_retreat') {
    const cutFrame = 150; // Cut to tiger crossing empty road at 5.0s (pause before "Moments later...")

    return (
      <div style={{ width: '100%', height: '100%', position: 'relative', overflow: 'hidden', backgroundColor: '#000' }}>
        {/* Shot 1: Woman leaving alert (plays from 0s to 5.0s) */}
        <Sequence from={0} durationInFrames={cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/09_woman_leaving_alert.mp4')}
            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
          />
        </Sequence>

        {/* Shot 2: Tiger crossing road (plays from frame 0 of video, in majestic slow motion to cover entire 11.9s without freezing) */}
        <Sequence from={cutFrame} durationInFrames={durationInFrames - cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/10_tiger_crossing_road.mp4')}
            playbackRate={0.65}
            style={{ width: '100%', height: '100%', objectFit: 'cover' }}
          />
        </Sequence>

        {/* Anamorphic letterbox bars */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 60, backgroundColor: '#000', zIndex: 40 }} />
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 60, backgroundColor: '#000', zIndex: 40 }} />

        {/* Location & Outcome Tag */}
        <div style={{ position: 'absolute', top: 80, left: 80, zIndex: 45 }}>
          <div style={{ color: frame >= cutFrame ? '#10B981' : '#F59E0B', fontSize: 16, fontWeight: 800, letterSpacing: '3px' }}>
            {frame >= cutFrame ? '✓ PEACEFUL CROSSING // ZERO FATALITIES' : '⚠️ WORKER EVACUATING AMBUSH CORRIDOR'}
          </div>
          <div style={{ color: '#fff', fontSize: 13, fontFamily: 'monospace' }}>
            📍 COEXIST SHIELD AI REAL-TIME INTERVENTION
          </div>
        </div>
      </div>
    );
  }

  // Part 4 Black Joke: Bear Prime Minister Inauguration Scene (Brief comedic punchline ~5s)
  if (item.id === 'p4_02_bear_pm_joke') {
    const flashOpacity = Math.sin(frame * 2.2) > 0.88 ? 0.4 : 0;

    return (
      <div style={{ width: '100%', height: '100%', position: 'relative', overflow: 'hidden' }}>
        <OffthreadVideo
          src={staticFile('cinematics/bear_prime_minister.mp4')}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
          }}
        />
        {/* Camera flash overlay */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            backgroundColor: '#ffffff',
            opacity: flashOpacity,
            pointerEvents: 'none',
          }}
        />
        {/* Breaking News Ticker */}
        <div
          style={{
            position: 'absolute',
            top: 60,
            left: 80,
            background: 'rgba(220, 38, 38, 0.95)',
            color: '#fff',
            padding: '10px 28px',
            fontSize: 24,
            fontWeight: 900,
            borderRadius: 8,
            boxShadow: '0 4px 20px rgba(0,0,0,0.7)',
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            letterSpacing: '1px',
          }}
        >
          <span>🔴</span> SPECIAL REPORT: TOKYO KANTEI
        </div>
      </div>
    );
  }

  // Part 4 Japan Crisis: Real Footage of Bears in Persimmon Trees and Residential Streets
  if (item.id === 'p4_01_japan_crisis') {
    const cutFrame = 180; // Cut at 6.0s: from persimmon tree to suburb street
    const isFirstHalf = frame < cutFrame;
    const chapterTitle = isFirstHalf
      ? 'V. ABANDONED PERSIMMON ATTRACTANT'
      : 'VI. RESIDENTIAL ENCROACHMENT';
    const locationTag = isFirstHalf
      ? 'IWATE & NAGANO VALLEYS // PRIVATE LAND'
      : 'SUBURBAN RESIDENTIAL DISTRICT // DUSK';

    return (
      <div
        style={{
          width: '100%',
          height: '100%',
          backgroundColor: '#000000',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Shot 1: Persimmon Tree bear foraging (plays from frame 0 to 180 = 6.0s) */}
        <Sequence from={0} durationInFrames={cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/05_japan_bear_persimmon.mp4')}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
            }}
          />
        </Sequence>

        {/* Shot 2: Suburb street bear encroachment (starts from frame 0 of video, plays 6.0s to 13.0s = 210 frames) */}
        <Sequence from={cutFrame} durationInFrames={durationInFrames - cutFrame} layout="none">
          <OffthreadVideo
            src={staticFile('cinematics/06_japan_bear_suburb.mp4')}
            style={{
              width: '100%',
              height: '100%',
              objectFit: 'cover',
            }}
          />
        </Sequence>

        {/* Cinematic film vignette */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: 'radial-gradient(circle, transparent 40%, rgba(0,0,0,0.7) 100%)',
            pointerEvents: 'none',
          }}
        />

        {/* Anamorphic letterbox bars */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 60, backgroundColor: '#000000', zIndex: 40 }} />
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 60, backgroundColor: '#000000', zIndex: 40 }} />

        {/* Location & Chapter Tag */}
        <div
          style={{
            position: 'absolute',
            top: 80,
            left: 80,
            display: 'flex',
            flexDirection: 'column',
            gap: 6,
            zIndex: 45,
          }}
        >
          <div
            style={{
              color: '#F59E0B',
              fontSize: 16,
              fontWeight: 800,
              letterSpacing: '3px',
              textTransform: 'uppercase',
            }}
          >
            {chapterTitle}
          </div>
          <div
            style={{
              color: 'rgba(255, 255, 255, 0.7)',
              fontSize: 12,
              fontWeight: 600,
              letterSpacing: '2px',
              fontFamily: 'monospace',
            }}
          >
            📍 JAPAN // {locationTag}
          </div>
        </div>
      </div>
    );
  }

  // Part 1: Real Cinematic Motion Video Clips
  if (item.id.startsWith('p1_')) {
    let videoFile = 'cinematics/01_tropical_mist.mp4';
    let chapterTitle = 'I. THE FRONTIER COLLISION';
    let locationTag = 'SOUTHEAST ASIA // FOREST-PLANTATION FRONTIER';

    if (item.id === 'p1_02_tiger') {
      videoFile = 'cinematics/02_tiger_stalk.mp4';
      chapterTitle = 'II. THE APEX AMBUSH';
      locationTag = 'SUMATRA, INDONESIA // LOWLAND PEAT FOREST';
    } else if (item.id === 'p1_03_elephant') {
      videoFile = 'cinematics/03_elephant_raid.mp4';
      chapterTitle = 'III. THE MIDNIGHT RAID';
      locationTag = 'PHETCHABURI, THAILAND // SUGARCANE CORRIDOR';
    } else if (item.id === 'p1_04_roadkill') {
      videoFile = 'cinematics/04_roadkill_rain.mp4';
      chapterTitle = 'IV. THE PATHOGEN SPILLOVER';
      locationTag = 'NATIONAL HIGHWAY 3219 // RAINFOREST ARTERY';
    }

    return (
      <div
        style={{
          width: '100%',
          height: '100%',
          backgroundColor: '#000000',
          position: 'relative',
          overflow: 'hidden',
        }}
      >
        {/* Fullscreen Cinematic Video */}
        <OffthreadVideo
          src={staticFile(videoFile)}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
          }}
        />

        {/* Film grain / dark vignette gradient overlay */}
        <div
          style={{
            position: 'absolute',
            inset: 0,
            background: 'radial-gradient(circle, transparent 40%, rgba(0,0,0,0.7) 100%)',
            pointerEvents: 'none',
          }}
        />

        {/* Anamorphic Letterbox Bars */}
        <div style={{ position: 'absolute', top: 0, left: 0, right: 0, height: 60, backgroundColor: '#000000', zIndex: 40 }} />
        <div style={{ position: 'absolute', bottom: 0, left: 0, right: 0, height: 60, backgroundColor: '#000000', zIndex: 40 }} />

        {/* Cinematic Location & Chapter Card */}
        <div
          style={{
            position: 'absolute',
            top: 80,
            left: 80,
            display: 'flex',
            flexDirection: 'column',
            gap: 6,
            zIndex: 45,
          }}
        >
          <div
            style={{
              color: '#38BDF8',
              fontSize: 16,
              fontWeight: 800,
              letterSpacing: '3px',
              textTransform: 'uppercase',
            }}
          >
            {chapterTitle}
          </div>
          <div
            style={{
              color: 'rgba(255, 255, 255, 0.7)',
              fontSize: 12,
              fontWeight: 600,
              letterSpacing: '2px',
              fontFamily: 'monospace',
            }}
          >
            📍 {locationTag}
          </div>
        </div>
      </div>
    );
  }

  // Part 2: Geofencing Core Explainer (What is a Geofence? -> Traditional Flaws -> CoExist Shield AI Innovation)
  if (item.id === 'p2_02_geofence') {
    return <GeofenceExplainer />;
  }

  // Part 2: Dynamic Geofence AI Algorithm Simulation
  if (item.id === 'p2_03_algorithm') {
    return <AlgorithmSimulation />;
  }

  // Part 2 & Part 5: High-Tech Branding / Architecture
  if (item.id.startsWith('p2_') || item.id.startsWith('p5_')) {
    return (
      <div
        style={{
          width: '100%',
          height: '100%',
          background: 'radial-gradient(circle at 50% 40%, #0F2744 0%, #060B14 100%)',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          position: 'relative',
        }}
      >
        <div
          style={{
            fontSize: 72,
            fontWeight: 900,
            background: 'linear-gradient(135deg, #38BDF8 0%, #22D3EE 50%, #34D399 100%)',
            WebkitBackgroundClip: 'text',
            WebkitTextFillColor: 'transparent',
            letterSpacing: '2px',
            marginBottom: 16,
          }}
        >
          CoExist Shield AI
        </div>
        <div
          style={{
            color: '#94A3B8',
            fontSize: 26,
            maxWidth: 1000,
            textAlign: 'center',
            lineHeight: 1.5,
          }}
        >
          Autonomous Predictive Geofencing & Ecological Early-Warning Platform
        </div>
      </div>
    );
  }

  // Part 3 & Part 4_04: System Dashboard Walkthrough (Simulation / Playwright capture)
  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        background: '#0B1320',
        padding: '50px 80px',
        boxSizing: 'border-box',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {/* Top App Header */}
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: '2px solid rgba(255,255,255,0.1)',
          paddingBottom: 20,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <span style={{ fontSize: 32 }}>🛡️</span>
          <span style={{ color: '#fff', fontSize: 26, fontWeight: 800 }}>CoExist Shield AI — Command Center</span>
        </div>
        <div style={{ display: 'flex', gap: 16 }}>
          <span style={{ background: '#10B981', color: '#fff', padding: '6px 14px', borderRadius: 6, fontWeight: 700, fontSize: 14 }}>● AI RUNNING</span>
          <span style={{ background: '#EF4444', color: '#fff', padding: '6px 14px', borderRadius: 6, fontWeight: 700, fontSize: 14 }}>ACTIVE GEOFENCES: 4</span>
        </div>
      </div>

      {/* Main Content Area: Dynamic layout based on scene */}
      {item.id === 'p3_04_geofence_alert' || item.id === 'p4_03_tora_reaction' || item.id === 'p4_04_japan_dx' ? (
        // Split Screen: Web-Admin on left, Ionic Mobile App mockup on right
        <div style={{ display: 'flex', gap: 30, marginTop: 25, flex: 1, alignItems: 'center' }}>
          {/* Left: Web-Admin Console */}
          <div
            style={{
              flex: 1.5,
              height: '100%',
              backgroundColor: '#0A0F1D',
              borderRadius: 16,
              border: '2px solid rgba(56, 189, 248, 0.4)',
              overflow: 'hidden',
              position: 'relative',
              boxShadow: '0 8px 32px rgba(0,0,0,0.6)',
            }}
          >
            <OffthreadVideo
              src={staticFile('screen-recordings/admin_demo.mp4')}
              style={{ width: '100%', height: '100%', objectFit: 'cover' }}
            />
            <div
              style={{
                position: 'absolute',
                top: 14,
                left: 14,
                backgroundColor: 'rgba(15, 23, 42, 0.85)',
                backdropFilter: 'blur(8px)',
                border: '1px solid rgba(56, 189, 248, 0.5)',
                color: '#38BDF8',
                fontWeight: 800,
                fontSize: 14,
                padding: '4px 12px',
                borderRadius: 6,
                letterSpacing: '1px',
              }}
            >
              🖥️ WEB-ADMIN OPERATIONS PORTAL
            </div>
          </div>

          {/* Right: Ionic Mobile Phone Mockup */}
          <div
            style={{
              width: 380,
              height: '100%',
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              position: 'relative',
            }}
          >
            <div
              style={{
                width: 340,
                height: 700,
                backgroundColor: '#000',
                borderRadius: 40,
                border: '6px solid #334155',
                boxShadow: '0 20px 50px rgba(0,0,0,0.9), 0 0 30px rgba(255, 107, 0, 0.4)',
                overflow: 'hidden',
                position: 'relative',
              }}
            >
              {/* Phone Camera Notch */}
              <div
                style={{
                  position: 'absolute',
                  top: 0,
                  left: '50%',
                  transform: 'translateX(-50%)',
                  width: 120,
                  height: 18,
                  backgroundColor: '#334155',
                  borderBottomLeftRadius: 12,
                  borderBottomRightRadius: 12,
                  zIndex: 20,
                }}
              />
              <OffthreadVideo
                src={staticFile('screen-recordings/ionic_mobile_demo.mp4')}
                style={{ width: '100%', height: '100%', objectFit: 'cover' }}
              />
            </div>
            <div
              style={{
                marginTop: 12,
                background: 'linear-gradient(90deg, #10B981, #059669)',
                color: '#fff',
                fontSize: 14,
                fontWeight: 800,
                padding: '4px 16px',
                borderRadius: 12,
                boxShadow: '0 4px 12px rgba(0,0,0,0.5)',
              }}
            >
              📱 IONIC MOBILE (FIELD WORKER & CITIZEN HUD)
            </div>
          </div>
        </div>
      ) : (
        // Standard Web-Admin View with Telemetry Sidebar
        <div style={{ display: 'flex', gap: 30, marginTop: 25, flex: 1 }}>
          <div
            style={{
              flex: 2,
              backgroundColor: '#0A0F1D',
              borderRadius: 16,
              border: '2px solid rgba(56, 189, 248, 0.4)',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
              position: 'relative',
              boxShadow: '0 8px 32px rgba(0,0,0,0.6)',
            }}
          >
            <OffthreadVideo
              src={staticFile('screen-recordings/admin_demo.mp4')}
              style={{
                width: '100%',
                height: '100%',
                objectFit: 'cover',
              }}
            />
            <div
              style={{
                position: 'absolute',
                top: 14,
                left: 14,
                backgroundColor: 'rgba(15, 23, 42, 0.85)',
                backdropFilter: 'blur(8px)',
                border: '1px solid rgba(56, 189, 248, 0.5)',
                color: '#38BDF8',
                fontWeight: 800,
                fontSize: 14,
                padding: '4px 12px',
                borderRadius: 6,
                letterSpacing: '1px',
              }}
            >
              🖥️ WEB-ADMIN: OPERATIONS COMMAND DASHBOARD (1920x1080)
            </div>
          </div>

          {/* Right Info Telemetry Panel */}
          <div
            style={{
              flex: 1,
              backgroundColor: '#0F172A',
              borderRadius: 16,
              padding: 24,
              display: 'flex',
              flexDirection: 'column',
              gap: 16,
              border: '1px solid rgba(255,255,255,0.08)',
            }}
          >
            <div style={{ color: '#38BDF8', fontWeight: 800, fontSize: 18, textTransform: 'uppercase' }}>
              Telemetry & Predictive Metrics
            </div>
            <div style={{ background: '#1E293B', padding: 16, borderRadius: 10, color: '#F8FAFC' }}>
              <div style={{ fontSize: 14, color: '#94A3B8' }}>Active JAPAC Geofences</div>
              <div style={{ fontSize: 26, fontWeight: 800, color: '#38BDF8' }}>6 Zones (2 Critical)</div>
            </div>
            <div style={{ background: '#1E293B', padding: 16, borderRadius: 10, color: '#F8FAFC' }}>
              <div style={{ fontSize: 14, color: '#94A3B8' }}>High-Risk Roadkill Carcasses</div>
              <div style={{ fontSize: 22, fontWeight: 700, color: '#F87171' }}>4 Pending Dispatch</div>
            </div>
            <div style={{ background: '#1E293B', padding: 16, borderRadius: 10, color: '#F8FAFC' }}>
              <div style={{ fontSize: 14, color: '#94A3B8' }}>Protected Mobile Fleet</div>
              <div style={{ fontSize: 20, color: '#34D399', fontWeight: 700 }}>1,420 Users Active</div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
