import React from 'react';
import { Img, interpolate, staticFile, useCurrentFrame } from 'remotion';

interface WipeAvatarProps {
  isSpeaking: boolean;
  avatarUrl?: string;
}

export const WipeAvatar: React.FC<WipeAvatarProps> = ({
  isSpeaking,
  avatarUrl = staticFile('avatars/ranger_tora.jpg'),
}) => {
  const frame = useCurrentFrame();

  // Subtle breathing / speaking bounce animation
  const bounce = isSpeaking ? Math.sin(frame / 2.5) * 6 : Math.sin(frame / 15) * 2;
  const rotation = isSpeaking ? Math.sin(frame / 4) * 1.5 : 0;
  
  // Neon cyber ring rotation
  const ringRotation = (frame * 1.5) % 360;

  return (
    <div
      style={{
        position: 'absolute',
        bottom: 40,
        right: 50,
        width: 320,
        height: 380,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'flex-end',
        zIndex: 60,
        pointerEvents: 'none',
        transform: `translateY(${-bounce}px) rotate(${rotation}deg)`,
      }}
    >
      {/* Outer Cyber Glow Ring */}
      <div
        style={{
          position: 'absolute',
          bottom: 20,
          width: 290,
          height: 290,
          borderRadius: '50%',
          border: '4px dashed rgba(255, 120, 0, 0.8)',
          boxShadow: isSpeaking
            ? '0 0 35px rgba(255, 107, 0, 0.9), inset 0 0 25px rgba(255, 107, 0, 0.5)'
            : '0 0 15px rgba(255, 107, 0, 0.4)',
          transform: `rotate(${ringRotation}deg)`,
          transition: 'box-shadow 0.2s ease',
        }}
      />

      {/* Character Image Container */}
      <div
        style={{
          width: 270,
          height: 270,
          borderRadius: '50%',
          overflow: 'hidden',
          border: '4px solid #FFA500',
          boxShadow: '0 8px 30px rgba(0,0,0,0.8)',
          backgroundColor: '#0c1626',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Img
          src={avatarUrl}
          style={{
            width: '100%',
            height: '100%',
            objectFit: 'cover',
            transform: 'scale(1.1) translateY(10px)',
          }}
        />
      </div>

      {/* Nameplate */}
      <div
        style={{
          marginTop: -25,
          zIndex: 70,
          background: 'linear-gradient(90deg, #FF5500 0%, #FF8C00 100%)',
          color: '#ffffff',
          fontWeight: 900,
          fontSize: 18,
          padding: '6px 24px',
          borderRadius: 20,
          border: '2px solid #FFFFFF',
          boxShadow: '0 4px 15px rgba(0,0,0,0.6)',
          letterSpacing: '1px',
        }}
      >
        RANGER TORA ⚡ LIVE
      </div>
    </div>
  );
};
