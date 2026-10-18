import React from 'react';
import { interpolate, useCurrentFrame } from 'remotion';

interface SubtitleProps {
  speaker: string;
  text: string;
}

export const Subtitles: React.FC<SubtitleProps> = ({ speaker, text }) => {
  const frame = useCurrentFrame();

  const isTora = speaker === 'Ranger Tora';
  const badgeColor = isTora ? 'linear-gradient(135deg, #FF6B00 0%, #FFA800 100%)' : 'linear-gradient(135deg, #1E3A8A 0%, #3B82F6 100%)';
  const badgeText = isTora ? '🐅 RANGER TORA' : '🎙️ NARRATION';

  const opacity = interpolate(frame, [0, 5], [0, 1], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  const translateY = interpolate(frame, [0, 5], [10, 0], {
    extrapolateLeft: 'clamp',
    extrapolateRight: 'clamp',
  });

  return (
    <div
      style={{
        position: 'absolute',
        bottom: 50,
        left: 0,
        right: 0,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '0 120px',
        opacity,
        transform: `translateY(${translateY}px)`,
        pointerEvents: 'none',
        zIndex: 80,
      }}
    >
      {/* Speaker Badge */}
      <div
        style={{
          background: badgeColor,
          color: '#ffffff',
          fontWeight: 800,
          fontSize: 22,
          padding: '6px 20px',
          borderRadius: 20,
          boxShadow: '0 4px 15px rgba(0,0,0,0.5)',
          textTransform: 'uppercase',
          letterSpacing: '1.5px',
          marginBottom: 10,
          alignSelf: 'center',
          border: '2px solid rgba(255,255,255,0.4)',
        }}
      >
        {badgeText}
      </div>

      {/* Main Subtitle Box */}
      <div
        style={{
          backgroundColor: 'rgba(10, 15, 25, 0.85)',
          backdropFilter: 'blur(12px)',
          border: isTora ? '2px solid rgba(255, 107, 0, 0.6)' : '2px solid rgba(59, 130, 246, 0.6)',
          borderRadius: 16,
          padding: '16px 36px',
          maxWidth: '1200px',
          textAlign: 'center',
          boxShadow: '0 10px 30px rgba(0,0,0,0.7)',
        }}
      >
        <p
          style={{
            margin: 0,
            color: '#FFFFFF',
            fontSize: 32,
            fontWeight: 700,
            lineHeight: 1.4,
            textShadow: '0 2px 4px rgba(0,0,0,0.8)',
            fontFamily: 'Inter, system-ui, -apple-system, sans-serif',
          }}
        >
          {text}
        </p>
      </div>
    </div>
  );
};
