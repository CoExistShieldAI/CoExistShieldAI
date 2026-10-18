import React from 'react';
import { interpolate, spring, useCurrentFrame, useVideoConfig } from 'remotion';

interface WearOSWatchModalProps {
  distance?: string;
  threat?: string;
}

export const WearOSWatchModal: React.FC<WearOSWatchModalProps> = ({
  distance = '150m',
  threat = 'SUMATRAN TIGER',
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Smooth pop-in animation using spring
  const scale = spring({
    frame,
    fps,
    config: {
      damping: 14,
      stiffness: 90,
    },
  });

  // Haptic pulse animation
  const pulse = Math.sin(frame / 2) * 8;
  const ringRotation = (frame * 2.5) % 360;

  return (
    <div
      style={{
        position: 'absolute',
        inset: 0,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 55,
        backgroundColor: 'rgba(3, 7, 18, 0.65)',
        backdropFilter: 'blur(10px)',
      }}
    >
      {/* WearOS Tag Header */}
      <div
        style={{
          marginBottom: 20,
          background: 'linear-gradient(90deg, #3B82F6, #1D4ED8)',
          color: '#ffffff',
          fontWeight: 900,
          fontSize: 20,
          padding: '8px 24px',
          borderRadius: 24,
          letterSpacing: '2px',
          boxShadow: '0 4px 20px rgba(0,0,0,0.6)',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
        }}
      >
        <span>⌚</span> WEAR OS TACTICAL HUD // WRIST HAPTIC ENGINE
      </div>

      {/* Main Circular Smartwatch Hardware Chassis */}
      <div
        style={{
          width: 520,
          height: 520,
          borderRadius: '50%',
          backgroundColor: '#0a0a0f',
          border: '14px solid #1f2937',
          boxShadow: `0 25px 60px rgba(0,0,0,0.95), 0 0 ${40 + pulse * 2}px rgba(239, 68, 68, 0.75)`,
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
          transform: `scale(${scale})`,
          overflow: 'hidden',
        }}
      >
        {/* Watch Crown button on right */}
        <div
          style={{
            position: 'absolute',
            right: -24,
            top: '50%',
            transform: 'translateY(-50%)',
            width: 14,
            height: 48,
            backgroundColor: '#374151',
            borderRadius: 4,
          }}
        />

        {/* Rotating Tactical Radar Ring */}
        <div
          style={{
            position: 'absolute',
            width: 460,
            height: 460,
            borderRadius: '50%',
            border: '3px dashed rgba(239, 68, 68, 0.8)',
            transform: `rotate(${ringRotation}deg)`,
            pointerEvents: 'none',
          }}
        />

        {/* Outer Circular Danger Arc Gauge */}
        <div
          style={{
            position: 'absolute',
            width: 420,
            height: 420,
            borderRadius: '50%',
            border: '6px solid rgba(239, 68, 68, 0.3)',
            borderTopColor: '#EF4444',
            borderRightColor: '#F97316',
            transform: 'rotate(-45deg)',
          }}
        />

        {/* Watch Face Screen Content */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            textAlign: 'center',
            padding: 30,
            zIndex: 10,
          }}
        >
          {/* Top Status */}
          <div
            style={{
              color: '#F87171',
              fontSize: 16,
              fontWeight: 800,
              letterSpacing: '3px',
              textTransform: 'uppercase',
              marginBottom: 4,
            }}
          >
            ● HAPTIC PULSE ACTIVE
          </div>

          {/* Alert Tier Badge */}
          <div
            style={{
              backgroundColor: '#EF4444',
              color: '#FFFFFF',
              fontWeight: 900,
              fontSize: 22,
              padding: '6px 22px',
              borderRadius: 20,
              boxShadow: '0 4px 15px rgba(239, 68, 68, 0.6)',
              marginBottom: 12,
              letterSpacing: '1px',
            }}
          >
            🚨 TIER 3 DANGER
          </div>

          {/* Proximity Distance (Big Metric) */}
          <div
            style={{
              fontSize: 72,
              fontWeight: 900,
              color: '#FFFFFF',
              lineHeight: 1,
              letterSpacing: '-1px',
              textShadow: '0 0 20px rgba(239, 68, 68, 0.9)',
            }}
          >
            {distance}
          </div>

          <div
            style={{
              color: '#FCA5A5',
              fontSize: 14,
              fontWeight: 700,
              letterSpacing: '2px',
              textTransform: 'uppercase',
              marginTop: 4,
            }}
          >
            ESTIMATED PROXIMITY
          </div>

          {/* Threat Species */}
          <div
            style={{
              marginTop: 12,
              color: '#FCD34D',
              fontSize: 20,
              fontWeight: 800,
              letterSpacing: '1px',
            }}
          >
            🐅 {threat}
          </div>

          {/* Action Recommendation */}
          <div
            style={{
              marginTop: 16,
              backgroundColor: 'rgba(255,255,255,0.15)',
              border: '1px solid rgba(255,255,255,0.3)',
              borderRadius: 14,
              padding: '6px 18px',
              color: '#FFFFFF',
              fontSize: 13,
              fontWeight: 700,
              letterSpacing: '1px',
            }}
          >
            ⚠️ TURN BACK // FOLLOW SAFE ROUTE
          </div>
        </div>
      </div>
    </div>
  );
};
