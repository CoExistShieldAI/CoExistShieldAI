import React from 'react';
import { interpolate, useCurrentFrame } from 'remotion';

export const AlgorithmSimulation: React.FC = () => {
  const frame = useCurrentFrame();

  // Pulse rates
  const pulse = (Math.sin(frame / 6) + 1) / 2;
  const radarAngle = (frame * 3.5) % 360;
  const scanY = (frame * 4) % 1080;

  // Real-time fluctuating simulated values
  const brixVal = (18.4 + Math.sin(frame / 20) * 0.8).toFixed(1);
  const windSpeed = Math.round(14 + Math.sin(frame / 15) * 3);
  const riskScore = interpolate(
    frame,
    [0, 150, 450, 700],
    [0.34, 0.72, 0.94, 0.96],
    { extrapolateRight: 'clamp' }
  );

  // SVG animated dash offset for data stream lines
  const dashOffset = -frame * 3.2;

  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        backgroundColor: '#030712',
        position: 'relative',
        overflow: 'hidden',
        color: '#fff',
        fontFamily: "'JetBrains Mono', 'Inter', monospace, sans-serif",
      }}
    >
      {/* Background Cyber Grid */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            linear-gradient(to right, rgba(56, 189, 248, 0.05) 1px, transparent 1px),
            linear-gradient(to bottom, rgba(56, 189, 248, 0.05) 1px, transparent 1px)
          `,
          backgroundSize: '40px 40px',
        }}
      />

      {/* Atmospheric radial glow */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background: 'radial-gradient(circle at 50% 35%, rgba(14, 165, 233, 0.14) 0%, transparent 70%)',
          pointerEvents: 'none',
        }}
      />

      {/* Laser Scanning Line */}
      <div
        style={{
          position: 'absolute',
          top: scanY,
          left: 0,
          right: 0,
          height: 2,
          background: 'linear-gradient(90deg, transparent, rgba(56, 189, 248, 0.8), rgba(34, 211, 238, 1), transparent)',
          boxShadow: '0 0 15px rgba(56, 189, 248, 0.8)',
          pointerEvents: 'none',
        }}
      />

      {/* TOP HEADER / SYSTEM STATUS BAR */}
      <div
        style={{
          position: 'absolute',
          top: 22,
          left: 60,
          right: 60,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: '1px solid rgba(56, 189, 248, 0.3)',
          paddingBottom: 10,
          zIndex: 20,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div
            style={{
              width: 14,
              height: 14,
              borderRadius: '50%',
              backgroundColor: '#10B981',
              boxShadow: `0 0 ${10 + pulse * 10}px #10B981`,
            }}
          />
          <div>
            <div style={{ fontSize: 21, fontWeight: 900, letterSpacing: '2px', color: '#38BDF8' }}>
              AUTONOMOUS CAUSAL INFERENCE ENGINE
            </div>
            <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.6)', letterSpacing: '1px' }}>
              DYNAMIC SPATIO-TEMPORAL PROBABILITY MODELING & GEOFENCE SYNTHESIS
            </div>
          </div>
        </div>

        <div style={{ display: 'flex', gap: 14, fontSize: 11.5 }}>
          <div style={{ background: 'rgba(15, 23, 42, 0.85)', border: '1px solid rgba(56, 189, 248, 0.3)', padding: '5px 12px', borderRadius: 6 }}>
            <span style={{ color: '#94A3B8' }}>PIPELINES: </span>
            <span style={{ color: '#34D399', fontWeight: 800 }}>5 LIVE STREAMS</span>
          </div>
          <div style={{ background: 'rgba(15, 23, 42, 0.85)', border: '1px solid rgba(56, 189, 248, 0.3)', padding: '5px 12px', borderRadius: 6 }}>
            <span style={{ color: '#94A3B8' }}>LATENCY: </span>
            <span style={{ color: '#38BDF8', fontWeight: 800 }}>38ms EDGE</span>
          </div>
          <div style={{ background: 'rgba(15, 23, 42, 0.85)', border: '1px solid rgba(239, 68, 68, 0.4)', padding: '5px 12px', borderRadius: 6 }}>
            <span style={{ color: '#94A3B8' }}>CALCULATED RISK: </span>
            <span style={{ color: '#EF4444', fontWeight: 800 }}>{(riskScore * 100).toFixed(1)}% (TIER 3 DANGER)</span>
          </div>
        </div>
      </div>

      {/* SVG DATA-FLOW BEAMS (Connecting Left -> Center -> Right) */}
      <svg
        style={{
          position: 'absolute',
          top: 85,
          left: 60,
          width: 1800,
          height: 640,
          pointerEvents: 'none',
          zIndex: 5,
        }}
      >
        <defs>
          <linearGradient id="streamGrad1" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="rgba(245, 158, 11, 0.9)" />
            <stop offset="100%" stopColor="rgba(56, 189, 248, 0.9)" />
          </linearGradient>
          <linearGradient id="streamGrad2" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="rgba(239, 68, 68, 0.9)" />
            <stop offset="100%" stopColor="rgba(56, 189, 248, 0.9)" />
          </linearGradient>
          <linearGradient id="outputGrad" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="rgba(56, 189, 248, 0.9)" />
            <stop offset="100%" stopColor="rgba(239, 68, 68, 1)" />
          </linearGradient>
        </defs>

        {/* Ingoing Curves from 5 Input Cards into Center Brain */}
        <path d="M 460 75 C 510 75, 520 230, 560 230" fill="none" stroke="url(#streamGrad1)" strokeWidth="2.5" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 460 170 C 510 170, 520 250, 560 250" fill="none" stroke="url(#streamGrad2)" strokeWidth="2.5" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 460 265 C 510 265, 520 270, 560 270" fill="none" stroke="rgba(6, 182, 212, 0.8)" strokeWidth="2.5" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 460 360 C 510 360, 520 290, 560 290" fill="none" stroke="rgba(236, 72, 153, 0.8)" strokeWidth="2" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 460 455 C 510 455, 520 310, 560 310" fill="none" stroke="rgba(16, 185, 129, 0.8)" strokeWidth="2" strokeDasharray="6 4" strokeDashoffset={dashOffset} />

        {/* Outgoing Curves from Center Brain to Column 3 Geofences */}
        <path d="M 1260 230 C 1290 230, 1300 78, 1330 78" fill="none" stroke="rgba(16, 185, 129, 0.85)" strokeWidth="2.5" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 1260 255 C 1290 255, 1300 185, 1330 185" fill="none" stroke="rgba(245, 158, 11, 0.85)" strokeWidth="2.5" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 1260 280 C 1290 280, 1300 295, 1330 295" fill="none" stroke="url(#outputGrad)" strokeWidth="3" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
        <path d="M 1260 305 C 1290 305, 1300 440, 1330 440" fill="none" stroke="rgba(56, 189, 248, 0.8)" strokeWidth="2" strokeDasharray="6 4" strokeDashoffset={dashOffset} />
      </svg>

      {/* THREE-COLUMN ALGORITHM ARCHITECTURE */}
      <div
        style={{
          position: 'absolute',
          top: 88,
          left: 60,
          right: 60,
          height: 605,
          display: 'grid',
          gridTemplateColumns: '460px 1fr 470px',
          gap: 30,
          zIndex: 10,
        }}
      >
        {/* COLUMN 1: MULTI-SOURCE INPUT DATA INGESTION */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              background: 'rgba(14, 165, 233, 0.15)',
              border: '1px solid rgba(56, 189, 248, 0.4)',
              padding: '5px 12px',
              borderRadius: 6,
            }}
          >
            <span style={{ fontSize: 11.5, fontWeight: 800, color: '#38BDF8', letterSpacing: '1px' }}>
              1. MULTI-SOURCE CAUSAL STREAMS
            </span>
            <span style={{ fontSize: 10, color: '#10B981', fontWeight: 700 }}>● INGESTING</span>
          </div>

          {/* Stream 1: Sugarcane Harvest Phenology */}
          <DataCard
            icon="🌾"
            title="CROP HARVEST PHENOLOGY"
            subtitle="Sugarcane & Jackfruit Ripening Plume"
            metric={`Brix ${brixVal}° (Peak Sugar) // Olfactory Index: 0.91`}
            status="HEC Attractant Active"
            color="#F59E0B"
            progress={0.88}
          />

          {/* Stream 2: Roadkill Carcass Cache */}
          <DataCard
            icon="🥩"
            title="ROADKILL & CARCASS CACHING"
            subtitle="72h Carnivore Return Window"
            metric="Fresh Bovine Carcass // Decay: <18h"
            status="Tiger Ambush Magnet (94% Prob)"
            color="#EF4444"
            progress={0.94}
          />

          {/* Stream 3: Meteorological Dynamics */}
          <DataCard
            icon="💨"
            title="DISPERSION & NOCTURNAL CLIMATE"
            subtitle="Wind Scent Vectors & Moon Illumination"
            metric={`Wind: ${windSpeed} km/h NE // Moon: New (0.01 Lux)`}
            status="Stealth Plume Drift Confirmed"
            color="#06B6D4"
            progress={0.76}
          />

          {/* Stream 4: NASA FIRMS Satellite */}
          <DataCard
            icon="🛰️"
            title="NASA FIRMS THERMAL & LOGGING"
            subtitle="Deforestation & Corridor Fragmentation"
            metric="2 Thermal Hotspots // Peat Border: 1.8km"
            status="Habitat Displacement Vector: S"
            color="#EC4899"
            progress={0.65}
          />

          {/* Stream 5: Citizen & Social Media NLP */}
          <DataCard
            icon="📱"
            title="CROWDSOURCED SIGHTINGS NLP"
            subtitle="Real-Time Traces & Audio Detections"
            metric="2 Scat & Claw Marks (LLM Conf: 96%)"
            status="Biological Ground Truth Verified"
            color="#10B981"
            progress={0.92}
          />
        </div>

        {/* COLUMN 2: GEOSPATIAL PROBABILITY INFERENCE BRAIN */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 8,
            background: 'rgba(10, 18, 36, 0.75)',
            border: '1px solid rgba(56, 189, 248, 0.35)',
            borderRadius: 12,
            padding: '10px 14px',
            position: 'relative',
            overflow: 'hidden',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              borderBottom: '1px solid rgba(56, 189, 248, 0.2)',
              paddingBottom: 5,
            }}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <span style={{ fontSize: 15 }}>🧠</span>
              <span style={{ fontSize: 12.5, fontWeight: 800, color: '#38BDF8', letterSpacing: '1px' }}>
                2. GEOSPATIAL PROBABILITY INFERENCE BRAIN
              </span>
            </div>
            <span style={{ fontSize: 10.5, color: '#F59E0B', fontFamily: 'monospace' }}>
              BAYESIAN SPATIAL TENSOR
            </span>
          </div>

          {/* Model Formula Ticker */}
          <div
            style={{
              background: 'rgba(15, 23, 42, 0.9)',
              border: '1px solid rgba(56, 189, 248, 0.2)',
              borderRadius: 6,
              padding: '5px 10px',
              fontSize: 10,
              fontFamily: 'monospace',
              color: '#94A3B8',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <span style={{ color: '#38BDF8', fontWeight: 800 }}>MODEL:</span>
            <span style={{ color: '#E2E8F0' }}>Risk(x,y,t) = σ( Σ w_i·Attractant_i + WindVector + CarcassDecay_72h )</span>
          </div>

          {/* Simulated 2D Topographic Risk Heatmap Radar */}
          <div
            style={{
              flex: 1,
              position: 'relative',
              borderRadius: 8,
              border: '1px solid rgba(56, 189, 248, 0.25)',
              backgroundColor: '#040914',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              overflow: 'hidden',
            }}
          >
            {/* Radar Circular Grid & Sweeper */}
            <svg
              style={{ position: 'absolute', inset: 0, width: '100%', height: '100%' }}
              viewBox="0 0 500 320"
            >
              <defs>
                <radialGradient id="riskHeat" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stopColor="rgba(239, 68, 68, 0.95)" />
                  <stop offset="35%" stopColor="rgba(245, 158, 11, 0.7)" />
                  <stop offset="70%" stopColor="rgba(16, 185, 129, 0.35)" />
                  <stop offset="100%" stopColor="transparent" />
                </radialGradient>
              </defs>

              {/* Grid Lines */}
              <line x1="250" y1="10" x2="250" y2="310" stroke="rgba(56, 189, 248, 0.2)" strokeDasharray="4 4" />
              <line x1="40" y1="155" x2="460" y2="155" stroke="rgba(56, 189, 248, 0.2)" strokeDasharray="4 4" />

              {/* Concentric distance rings */}
              <circle cx="250" cy="155" r="130" fill="none" stroke="rgba(56, 189, 248, 0.25)" strokeWidth="1" />
              <circle cx="250" cy="155" r="88" fill="none" stroke="rgba(56, 189, 248, 0.35)" strokeWidth="1" strokeDasharray="6 3" />
              <circle cx="250" cy="155" r="45" fill="none" stroke="rgba(56, 189, 248, 0.45)" strokeWidth="1" />

              {/* Dynamic Risk Heatmap Blob (Morphs and shifts with wind vector) */}
              <ellipse
                cx={250 + Math.sin(frame / 20) * 12}
                cy={150 + Math.cos(frame / 25) * 8}
                rx={82 + Math.sin(frame / 10) * 6}
                ry={66 + Math.cos(frame / 12) * 5}
                fill="url(#riskHeat)"
                style={{
                  filter: 'blur(7px)',
                  transformOrigin: '250px 155px',
                  transform: `rotate(${Math.sin(frame / 30) * 10}deg)`,
                }}
              />

              {/* Scent Plume Dispersion Stream Vector */}
              <path
                d="M 120 235 Q 190 195, 250 150 T 380 90"
                fill="none"
                stroke="rgba(6, 182, 212, 0.8)"
                strokeWidth="2.5"
                strokeDasharray="8 6"
                strokeDashoffset={dashOffset}
              />
              <text x="80" y="255" fill="#06B6D4" fontSize="10.5" fontWeight="700">
                💨 DOWNWIND SCENT DISPERSION PLUME
              </text>

              {/* Predicted Carnivore Movement Vector */}
              <line
                x1="250"
                y1="150"
                x2={310 + Math.cos(frame / 15) * 8}
                y2={110 + Math.sin(frame / 15) * 8}
                stroke="#EF4444"
                strokeWidth="3"
              />
              <circle cx="250" cy="150" r="5" fill="#EF4444" />
              <text x="260" y="170" fill="#EF4444" fontSize="10.5" fontWeight="800">
                🎯 LIKELY AMBUSH INTERCEPT
              </text>

              {/* Rotating Radar Scanner Sweep */}
              <line
                x1="250"
                y1="155"
                x2={250 + 130 * Math.cos((radarAngle * Math.PI) / 180)}
                y2={155 + 130 * Math.sin((radarAngle * Math.PI) / 180)}
                stroke="rgba(56, 189, 248, 0.7)"
                strokeWidth="2"
              />
            </svg>

            {/* Radar Coordinates HUD Overlays */}
            <div style={{ position: 'absolute', top: 8, left: 12, fontSize: 10, color: '#38BDF8', fontFamily: 'monospace' }}>
              LAT: 0°54'12"N // LNG: 101°26'34"E
            </div>
            <div style={{ position: 'absolute', top: 8, right: 12, fontSize: 10, color: '#10B981', fontFamily: 'monospace' }}>
              GRID: 50m²
            </div>

            {/* In-Radar Risk Convergence HUD */}
            <div
              style={{
                position: 'absolute',
                bottom: 8,
                left: 12,
                right: 12,
                backgroundColor: 'rgba(5, 10, 20, 0.92)',
                border: '1px solid rgba(239, 68, 68, 0.5)',
                borderRadius: 6,
                padding: '5px 12px',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}
            >
              <div>
                <div style={{ color: '#EF4444', fontSize: 10.5, fontWeight: 900, letterSpacing: '1px' }}>
                  🚨 SPATIO-TEMPORAL PROBABILITY CONVERGENCE
                </div>
                <div style={{ color: '#94A3B8', fontSize: 9 }}>
                  Priors: 72h Carcass + Night Downwind Plume + Sugarcane Border
                </div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ color: '#fff', fontSize: 15, fontWeight: 900, fontFamily: 'monospace' }}>
                  P = {(riskScore * 0.98).toFixed(3)}
                </div>
                <div style={{ color: '#34D399', fontSize: 8.5 }}>BAYESIAN OPTIMAL</div>
              </div>
            </div>
          </div>
        </div>

        {/* COLUMN 3: ADAPTIVE GEOFENCE GENERATION & REAL-TIME DISPATCH */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 7 }}>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              background: 'rgba(239, 68, 68, 0.15)',
              border: '1px solid rgba(239, 68, 68, 0.4)',
              padding: '5px 12px',
              borderRadius: 6,
            }}
          >
            <span style={{ fontSize: 11.5, fontWeight: 800, color: '#EF4444', letterSpacing: '1px' }}>
              3. DYNAMIC GEOFENCE SYNTHESIS
            </span>
            <span style={{ fontSize: 10, color: '#10B981', fontWeight: 700 }}>● REAL-TIME DISPATCH</span>
          </div>

          {/* Tier 1 Geofence: Notice Buffer */}
          <GeofenceCard
            tier="TIER 1 : NOTICE BUFFER"
            radius="RADIUS: 3,500m"
            desc="Broad agrarian buffer. Informs harvesters to coordinate daytime cutting."
            status="Harvest Coordination Active"
            color="#10B981"
            isActive={true}
          />

          {/* Tier 2 Geofence: Warning Corridor */}
          <GeofenceCard
            tier="TIER 2 : WARNING CORRIDOR"
            radius="RADIUS: 1,800m"
            desc="Scent dispersion interception zone. Triggers acoustic deterrents & detour prompts."
            status="Detour Advisories Dispatched"
            color="#F59E0B"
            isActive={true}
          />

          {/* Tier 3 Geofence: Critical Hazard Polygon */}
          <GeofenceCard
            tier="TIER 3 : CRITICAL DANGER POLYGON"
            radius="RADIUS: 450m (ACTIVE LOCK)"
            desc="Immediate ambush & kill cache boundary. Mandatory evacuation sirens & wrist alerts."
            status="🚨 IMMEDIATE EVACUATION LOCK"
            color="#EF4444"
            isActive={true}
            isFlashing={true}
          />

          {/* Downstream Device Dispatch Status Hub */}
          <div
            style={{
              background: 'rgba(10, 18, 36, 0.85)',
              border: '1px solid rgba(56, 189, 248, 0.3)',
              borderRadius: 8,
              padding: '8px 12px',
              display: 'flex',
              flexDirection: 'column',
              gap: 5,
            }}
          >
            <div style={{ fontSize: 10.5, fontWeight: 800, color: '#38BDF8', letterSpacing: '1px' }}>
              📡 INSTANT CROSS-PLATFORM DISPATCH:
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 10 }}>
              <span style={{ color: '#E2E8F0' }}>🖥️ Web-Admin Operations Portal</span>
              <span style={{ color: '#10B981', fontWeight: 800 }}>✓ SYNCED (WebSocket 10Hz)</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 10 }}>
              <span style={{ color: '#E2E8F0' }}>📱 Ionic Field Mobile HUD</span>
              <span style={{ color: '#F59E0B', fontWeight: 800 }}>✓ TIER 3 PUSH DELIVERED</span>
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: 10 }}>
              <span style={{ color: '#E2E8F0' }}>⌚ Wear OS Smartwatch</span>
              <span style={{ color: '#EF4444', fontWeight: 800 }}>
                🚨 SOS HAPTIC ACTIVE
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Sub-component: Multi-Source Data Card
interface DataCardProps {
  icon: string;
  title: string;
  subtitle: string;
  metric: string;
  status: string;
  color: string;
  progress: number;
}

const DataCard: React.FC<DataCardProps> = ({
  icon,
  title,
  subtitle,
  metric,
  status,
  color,
  progress,
}) => {
  return (
    <div
      style={{
        background: 'rgba(10, 18, 36, 0.75)',
        border: `1px solid ${color}44`,
        borderLeft: `4px solid ${color}`,
        borderRadius: 6,
        padding: '6px 10px',
        display: 'flex',
        flexDirection: 'column',
        gap: 2,
        position: 'relative',
        boxShadow: `0 2px 10px rgba(0,0,0,0.5)`,
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ fontSize: 12 }}>{icon}</span>
          <span style={{ fontSize: 10, fontWeight: 800, color: '#fff', letterSpacing: '0.5px' }}>
            {title}
          </span>
        </div>
        <span style={{ fontSize: 8.5, color, fontWeight: 700, fontFamily: 'monospace' }}>
          {status}
        </span>
      </div>

      <div style={{ fontSize: 9, color: 'rgba(255,255,255,0.6)' }}>
        {subtitle}
      </div>

      <div style={{ fontSize: 9, color: '#38BDF8', fontFamily: 'monospace', fontWeight: 600 }}>
        {metric}
      </div>

      {/* Mini Progress Bar */}
      <div
        style={{
          width: '100%',
          height: 3,
          backgroundColor: 'rgba(255,255,255,0.1)',
          borderRadius: 2,
          marginTop: 1,
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            width: `${progress * 100}%`,
            height: '100%',
            backgroundColor: color,
            borderRadius: 2,
            boxShadow: `0 0 6px ${color}`,
          }}
        />
      </div>
    </div>
  );
};

// Sub-component: Geofence Tier Card
interface GeofenceCardProps {
  tier: string;
  radius: string;
  desc: string;
  status: string;
  color: string;
  isActive: boolean;
  isFlashing?: boolean;
}

const GeofenceCard: React.FC<GeofenceCardProps> = ({
  tier,
  radius,
  desc,
  status,
  color,
  isActive,
  isFlashing = false,
}) => {
  const frame = useCurrentFrame();
  const flash = isFlashing && Math.sin(frame / 4) > 0;

  return (
    <div
      style={{
        background: flash ? `${color}22` : 'rgba(10, 18, 36, 0.75)',
        border: `1px solid ${flash ? color : `${color}55`}`,
        borderLeft: `4px solid ${color}`,
        borderRadius: 6,
        padding: '7px 11px',
        display: 'flex',
        flexDirection: 'column',
        gap: 2.5,
        position: 'relative',
        transition: 'all 0.1s ease',
        boxShadow: flash ? `0 0 15px ${color}66` : 'none',
      }}
    >
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: 10.5, fontWeight: 900, color, letterSpacing: '0.5px' }}>
          {tier}
        </span>
        <span style={{ fontSize: 9.5, fontWeight: 800, color: '#fff', fontFamily: 'monospace' }}>
          {radius}
        </span>
      </div>

      <div style={{ fontSize: 9, color: 'rgba(255,255,255,0.7)', lineHeight: 1.3 }}>
        {desc}
      </div>

      <div
        style={{
          fontSize: 9,
          color,
          fontWeight: 800,
          fontFamily: 'monospace',
          marginTop: 1,
          display: 'flex',
          alignItems: 'center',
          gap: 5,
        }}
      >
        <span>●</span> {status}
      </div>
    </div>
  );
};
