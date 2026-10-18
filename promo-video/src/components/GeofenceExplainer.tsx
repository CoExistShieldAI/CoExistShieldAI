import React from 'react';
import { interpolate, useCurrentFrame } from 'remotion';

export const GeofenceExplainer: React.FC = () => {
  const frame = useCurrentFrame();

  // Phase calculation based on speech:
  // Phase 1 (0 to 270 frames / 0s - 9s): "What is Geofencing? Invisible GPS fence"
  // Phase 2 (270 to 520 frames / 9s - 17.3s): "The Traditional Flaw: Static fences fail"
  // Phase 3 (520 to end / 17.3s - 31.6s): "CoExist Shield AI: Dynamic Predictive Geofence"
  const phase = frame < 270 ? 1 : frame < 520 ? 2 : 3;

  // Global animations
  const pulse = (Math.sin(frame / 6) + 1) / 2;

  // Phase 1 Animations: Worker crossing virtual fence
  const workerX1 = interpolate(frame, [0, 240], [120, 680], { extrapolateRight: 'clamp' });
  const isInsideP1 = workerX1 >= 420;
  const breachPulse = isInsideP1 ? (Math.sin((frame - 120) / 3) + 1) / 2 : 0;

  // Phase 2 Animations: Tiger moving freely outside static fence
  const tigerX2 = interpolate(frame, [270, 520], [160, 800], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });
  const tigerY2 = interpolate(frame, [270, 395, 520], [380, 430, 370], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });

  // Phase 3 Animations: Tiger with dynamic predictive 3-tier geofence moving together
  const tigerX3 = interpolate(frame, [520, 947], [220, 780], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });
  const tigerY3 = interpolate(frame, [520, 730, 947], [320, 270, 290], { extrapolateLeft: 'clamp', extrapolateRight: 'clamp' });

  // Predictive vector angle
  const predAngle = Math.atan2(290 - 320, 780 - 220) * (180 / Math.PI);

  return (
    <div
      style={{
        width: '100%',
        height: '100%',
        backgroundColor: '#030712',
        position: 'relative',
        overflow: 'hidden',
        color: '#fff',
        fontFamily: "'JetBrains Mono', 'Inter', -apple-system, sans-serif",
      }}
    >
      {/* Background Cyber-GIS Tactical Grid */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          backgroundImage: `
            linear-gradient(to right, rgba(56, 189, 248, 0.06) 1px, transparent 1px),
            linear-gradient(to bottom, rgba(56, 189, 248, 0.06) 1px, transparent 1px)
          `,
          backgroundSize: '44px 44px',
        }}
      />

      {/* Atmospheric Radial Vignette */}
      <div
        style={{
          position: 'absolute',
          inset: 0,
          background: 'radial-gradient(circle at 50% 40%, rgba(14, 165, 233, 0.12) 0%, rgba(3, 7, 18, 0.88) 85%)',
          pointerEvents: 'none',
        }}
      />

      {/* TOP HEADER / HUD NAVIGATION BAR */}
      <div
        style={{
          position: 'absolute',
          top: 20,
          left: 60,
          right: 60,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          borderBottom: '1px solid rgba(56, 189, 248, 0.3)',
          paddingBottom: 12,
          zIndex: 30,
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div
            style={{
              padding: '4px 12px',
              borderRadius: 6,
              background: 'linear-gradient(135deg, #0284C7, #0369A1)',
              fontSize: 12,
              fontWeight: 800,
              letterSpacing: '1.5px',
            }}
          >
            CORE ARCHITECTURE
          </div>
          <div style={{ fontSize: 22, fontWeight: 900, letterSpacing: '1px', color: '#F8FAFC' }}>
            WHAT IS GEOFENCING? <span style={{ color: '#38BDF8', fontSize: 18 }}>// ジオフェンシングの原理と革新</span>
          </div>
        </div>

        {/* 3-Step Interactive Breadcrumbs */}
        <div style={{ display: 'flex', gap: 10 }}>
          <div
            style={{
              padding: '6px 14px',
              borderRadius: 8,
              fontSize: 12,
              fontWeight: 800,
              border: phase === 1 ? '2px solid #38BDF8' : '1px solid rgba(255,255,255,0.15)',
              backgroundColor: phase === 1 ? 'rgba(56, 189, 248, 0.25)' : 'rgba(15, 23, 42, 0.6)',
              color: phase === 1 ? '#38BDF8' : '#64748B',
              transition: 'all 0.3s ease',
            }}
          >
            01. CONCEPT (仮想境界線)
          </div>
          <div
            style={{
              padding: '6px 14px',
              borderRadius: 8,
              fontSize: 12,
              fontWeight: 800,
              border: phase === 2 ? '2px solid #EF4444' : '1px solid rgba(255,255,255,0.15)',
              backgroundColor: phase === 2 ? 'rgba(239, 68, 68, 0.25)' : 'rgba(15, 23, 42, 0.6)',
              color: phase === 2 ? '#EF4444' : '#64748B',
              transition: 'all 0.3s ease',
            }}
          >
            02. TRADITIONAL FLAW (固定式の限界)
          </div>
          <div
            style={{
              padding: '6px 14px',
              borderRadius: 8,
              fontSize: 12,
              fontWeight: 800,
              border: phase === 3 ? '2px solid #10B981' : '1px solid rgba(255,255,255,0.15)',
              backgroundColor: phase === 3 ? 'rgba(16, 185, 129, 0.25)' : 'rgba(15, 23, 42, 0.6)',
              color: phase === 3 ? '#10B981' : '#64748B',
              transition: 'all 0.3s ease',
            }}
          >
            03. DYNAMIC AI SHIELD (動的予測型)
          </div>
        </div>
      </div>

      {/* =========================================================================
          PHASE 1: WHAT IS A GEOFENCE? (GPS-based invisible virtual boundary)
         ========================================================================= */}
      {phase === 1 && (
        <>
          {/* Phase 1 Text Header (Above Map) */}
          <div
            style={{
              position: 'absolute',
              top: 74,
              left: 60,
              right: 60,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              zIndex: 20,
            }}
          >
            <div>
              <div style={{ color: '#38BDF8', fontSize: 13, fontWeight: 800, letterSpacing: '2px' }}>
                📍 DEFINITION: GEOFENCING (ジオフェンスとは？)
              </div>
              <div style={{ fontSize: 23, fontWeight: 800, color: '#fff', marginTop: 2 }}>
                An <span style={{ color: '#38BDF8' }}>invisible virtual perimeter</span> created around a real-world area using GPS coordinates.
              </div>
              <div style={{ fontSize: 13, color: '#94A3B8', marginTop: 2 }}>
                物理的な柵（金網や電気柵）を建設することなく、衛星測位によって地図上に設定する目に見えないデジタル境界線。
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'rgba(15, 23, 42, 0.8)', padding: '6px 14px', borderRadius: 8, border: '1px solid rgba(56, 189, 248, 0.3)' }}>
              <span style={{ fontSize: 24 }}>🛰️</span>
              <div>
                <div style={{ fontSize: 11, fontWeight: 800, color: '#38BDF8' }}>GNSS CONSTELLATION</div>
                <div style={{ fontSize: 10, color: '#64748B' }}>REAL-TIME GPS TELEMETRY</div>
              </div>
            </div>
          </div>

          {/* Map Ground Plane Simulation (Positioned clearly above Subtitles!) */}
          <div
            style={{
              position: 'absolute',
              top: 170,
              left: 60,
              right: 60,
              height: 570,
              backgroundColor: 'rgba(10, 18, 32, 0.85)',
              borderRadius: 18,
              border: '1px solid rgba(56, 189, 248, 0.35)',
              boxShadow: '0 12px 40px rgba(0,0,0,0.8)',
              overflow: 'hidden',
              zIndex: 10,
            }}
          >
            {/* Dirt road / commute path */}
            <div
              style={{
                position: 'absolute',
                top: '52%',
                left: 0,
                right: 0,
                height: 70,
                transform: 'translateY(-50%)',
                backgroundColor: 'rgba(51, 65, 85, 0.4)',
                borderTop: '2px dashed rgba(148, 163, 184, 0.3)',
                borderBottom: '2px dashed rgba(148, 163, 184, 0.3)',
                display: 'flex',
                alignItems: 'center',
                paddingLeft: 30,
              }}
            >
              <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', letterSpacing: '2px' }}>
                COMMUTING ROADWAY // FIELD ACCESS CORRIDOR
              </span>
            </div>

            {/* Virtual Geofence Zone (Cyan circle) */}
            <div
              style={{
                position: 'absolute',
                top: '52%',
                left: 560,
                transform: 'translate(-50%, -50%)',
                width: 320,
                height: 320,
                borderRadius: '50%',
                border: '3px dashed #38BDF8',
                backgroundColor: 'rgba(14, 165, 233, 0.12)',
                boxShadow: `0 0 ${20 + pulse * 20}px rgba(56, 189, 248, 0.35)`,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <div
                style={{
                  background: 'rgba(15, 23, 42, 0.9)',
                  padding: '5px 12px',
                  borderRadius: 6,
                  border: '1px solid #38BDF8',
                  fontSize: 13,
                  fontWeight: 800,
                  color: '#38BDF8',
                  letterSpacing: '1px',
                }}
              >
                🛡️ VIRTUAL GEOFENCE
              </div>
              <div style={{ fontSize: 11, color: '#94A3B8', marginTop: 4 }}>
                RADIUS: 250m // NO PHYSICAL POSTS
              </div>
              <div style={{ fontSize: 10, color: '#64748B', marginTop: 2 }}>
                LAT: 08°24'N / LON: 99°58'E
              </div>
            </div>

            {/* Moving Field Worker */}
            <div
              style={{
                position: 'absolute',
                top: '52%',
                left: workerX1,
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                zIndex: 20,
              }}
            >
              {isInsideP1 && (
                <div
                  style={{
                    position: 'absolute',
                    top: -12,
                    width: 80,
                    height: 80,
                    borderRadius: '50%',
                    border: '3px solid #EF4444',
                    backgroundColor: 'rgba(239, 68, 68, 0.25)',
                    transform: `scale(${1 + breachPulse * 0.8})`,
                    opacity: 1 - breachPulse,
                    pointerEvents: 'none',
                  }}
                />
              )}
              <div
                style={{
                  width: 52,
                  height: 52,
                  borderRadius: '50%',
                  backgroundColor: isInsideP1 ? '#EF4444' : '#0284C7',
                  border: '3px solid #fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 26,
                  boxShadow: '0 4px 15px rgba(0,0,0,0.6)',
                }}
              >
                🚶
              </div>
              <div
                style={{
                  marginTop: 6,
                  background: isInsideP1 ? '#EF4444' : '#1E293B',
                  color: '#fff',
                  padding: '2px 8px',
                  borderRadius: 4,
                  fontSize: 11,
                  fontWeight: 700,
                  whiteSpace: 'nowrap',
                }}
              >
                {isInsideP1 ? '🚨 BOUNDARY BREACH!' : '📱 FIELD WORKER'}
              </div>
            </div>

            {/* Pop-up explanation card on breach */}
            {isInsideP1 && (
              <div
                style={{
                  position: 'absolute',
                  top: 24,
                  right: 30,
                  backgroundColor: 'rgba(15, 23, 42, 0.95)',
                  border: '2px solid #EF4444',
                  borderRadius: 12,
                  padding: '14px 20px',
                  maxWidth: 380,
                  boxShadow: '0 10px 30px rgba(239, 68, 68, 0.3)',
                  zIndex: 30,
                }}
              >
                <div style={{ color: '#EF4444', fontWeight: 900, fontSize: 15, display: 'flex', gap: 8, alignItems: 'center' }}>
                  <span>⚡</span> AUTOMATED TRIGGER FIRED
                </div>
                <div style={{ color: '#fff', fontSize: 13, fontWeight: 700, marginTop: 6, lineHeight: 1.4 }}>
                  Enters virtual zone ➔ Instant WearOS & Phone Haptic Alert
                </div>
                <div style={{ color: '#94A3B8', fontSize: 11, marginTop: 4 }}>
                  端末が仮想境界線に接触した瞬間、触覚振動・警告音・安全退避指示が自動配信されます。
                </div>
              </div>
            )}
          </div>
        </>
      )}

      {/* =========================================================================
          PHASE 2: THE TRADITIONAL FLAW (Static & Rigid boundaries fail with wildlife)
         ========================================================================= */}
      {phase === 2 && (
        <>
          {/* Phase 2 Text Header */}
          <div
            style={{
              position: 'absolute',
              top: 74,
              left: 60,
              right: 60,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              zIndex: 20,
            }}
          >
            <div>
              <div style={{ color: '#EF4444', fontSize: 13, fontWeight: 800, letterSpacing: '2px' }}>
                ❌ THE LIMITATION OF TRADITIONAL GEOFENCING (従来の限界)
              </div>
              <div style={{ fontSize: 23, fontWeight: 800, color: '#fff', marginTop: 2 }}>
                Traditional geofences are <span style={{ color: '#EF4444' }}>static and rigid</span>. But wildlife roams dynamically!
              </div>
              <div style={{ fontSize: 13, color: '#94A3B8', marginTop: 2 }}>
                従来のジオフェンスは「固定された円や四角形」。動物は固定された枠の外を自在に歩き回るため、枠外での予期せぬ遭遇事故を防げません。
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'rgba(239, 68, 68, 0.15)', padding: '6px 14px', borderRadius: 8, border: '1px solid rgba(239, 68, 68, 0.4)' }}>
              <span style={{ fontSize: 22 }}>⚠️</span>
              <div>
                <div style={{ fontSize: 11, fontWeight: 800, color: '#EF4444' }}>BLINDSPOT RISK</div>
                <div style={{ fontSize: 10, color: '#CBD5E1' }}>UNPREDICTED AMBUSH CORRIDOR</div>
              </div>
            </div>
          </div>

          {/* Map Ground Plane Simulation */}
          <div
            style={{
              position: 'absolute',
              top: 170,
              left: 60,
              right: 60,
              height: 570,
              backgroundColor: 'rgba(15, 23, 42, 0.85)',
              borderRadius: 18,
              border: '1px solid rgba(239, 68, 68, 0.4)',
              boxShadow: '0 12px 40px rgba(0,0,0,0.8)',
              overflow: 'hidden',
              zIndex: 10,
            }}
          >
            {/* Dirt road */}
            <div
              style={{
                position: 'absolute',
                top: '52%',
                left: 0,
                right: 0,
                height: 70,
                transform: 'translateY(-50%)',
                backgroundColor: 'rgba(51, 65, 85, 0.4)',
                borderTop: '2px dashed rgba(148, 163, 184, 0.3)',
                borderBottom: '2px dashed rgba(148, 163, 184, 0.3)',
              }}
            />

            {/* Static Rigid Geofence (Left side, gray) */}
            <div
              style={{
                position: 'absolute',
                top: 60,
                left: 80,
                width: 250,
                height: 250,
                borderRadius: '50%',
                border: '3px solid rgba(148, 163, 184, 0.4)',
                backgroundColor: 'rgba(71, 85, 105, 0.15)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <div
                style={{
                  background: '#334155',
                  padding: '4px 10px',
                  borderRadius: 6,
                  fontSize: 12,
                  fontWeight: 800,
                  color: '#94A3B8',
                }}
              >
                🔒 STATIC FIXED GEOFENCE
              </div>
              <div style={{ fontSize: 11, color: '#64748B', marginTop: 4 }}>CANNOT MOVE (固定式)</div>
              <div style={{ fontSize: 16, marginTop: 8 }}>🏘️ Village Reserve</div>
            </div>

            {/* Moving Tiger roams OUTSIDE the fixed geofence! */}
            <div
              style={{
                position: 'absolute',
                left: tigerX2,
                top: tigerY2,
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                zIndex: 20,
              }}
            >
              <div
                style={{
                  width: 56,
                  height: 56,
                  borderRadius: '50%',
                  backgroundColor: '#EA580C',
                  border: '3px solid #FDBA74',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 28,
                  boxShadow: '0 4px 20px rgba(234, 88, 12, 0.7)',
                }}
              >
                🐅
              </div>
              <div
                style={{
                  marginTop: 6,
                  background: 'rgba(234, 88, 12, 0.95)',
                  color: '#fff',
                  padding: '2px 8px',
                  borderRadius: 4,
                  fontSize: 11,
                  fontWeight: 800,
                  whiteSpace: 'nowrap',
                }}
              >
                ROAMING PREDATOR (動く野生動物)
              </div>
            </div>

            {/* Field Worker on the road */}
            <div
              style={{
                position: 'absolute',
                left: 880,
                top: '52%',
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                zIndex: 20,
              }}
            >
              <div
                style={{
                  width: 48,
                  height: 48,
                  borderRadius: '50%',
                  backgroundColor: '#0284C7',
                  border: '3px solid #fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 22,
                }}
              >
                🚶
              </div>
              <div
                style={{
                  marginTop: 6,
                  background: '#1E293B',
                  color: '#94A3B8',
                  padding: '2px 8px',
                  borderRadius: 4,
                  fontSize: 11,
                  fontWeight: 700,
                }}
              >
                UNAWARE WORKER
              </div>
            </div>

            {/* Blindspot Warning Callout */}
            <div
              style={{
                position: 'absolute',
                top: 24,
                right: 30,
                backgroundColor: 'rgba(239, 68, 68, 0.15)',
                border: '2px solid #EF4444',
                borderRadius: 12,
                padding: '14px 20px',
                maxWidth: 400,
                boxShadow: '0 8px 30px rgba(0,0,0,0.6)',
              }}
            >
              <div style={{ color: '#EF4444', fontSize: 15, fontWeight: 900, display: 'flex', alignItems: 'center', gap: 8 }}>
                <span>⚠️</span> CRITICAL SENSOR BLINDSPOT
              </div>
              <div style={{ color: '#fff', fontSize: 13, fontWeight: 700, marginTop: 6, lineHeight: 1.4 }}>
                Static fences leave workers unprotected in corridors outside the box!
              </div>
              <div style={{ color: '#CBD5E1', fontSize: 11, marginTop: 4 }}>
                虎や象は農地や作物の匂いに引かれて常にルートを変えます。固定のジオフェンスでは、枠外で起こる鉢合わせ衝突を防げません。
              </div>
            </div>
          </div>
        </>
      )}

      {/* =========================================================================
          PHASE 3: COEXIST SHIELD AI - DYNAMIC PREDICTIVE GEOFENCING
         ========================================================================= */}
      {phase === 3 && (
        <>
          {/* Phase 3 Text Header */}
          <div
            style={{
              position: 'absolute',
              top: 74,
              left: 60,
              right: 60,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'flex-start',
              zIndex: 20,
            }}
          >
            <div>
              <div style={{ color: '#10B981', fontSize: 13, fontWeight: 800, letterSpacing: '2px' }}>
                🛡️ THE BREAKTHROUGH: DYNAMIC PREDICTIVE GEOFENCING
              </div>
              <div style={{ fontSize: 23, fontWeight: 800, color: '#fff', marginTop: 2 }}>
                An <span style={{ color: '#10B981' }}>autonomous AI safety bubble</span> that moves with wildlife and forecasts trajectory!
              </div>
              <div style={{ fontSize: 13, color: '#94A3B8', marginTop: 2 }}>
                CoExist Shield AIはジオフェンスを「動的（Dynamic）」かつ「予測型（Predictive）」へ進化。動物の周囲にリアルタイムで3段階の安全シールドを展開します。
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, background: 'rgba(16, 185, 129, 0.15)', padding: '6px 14px', borderRadius: 8, border: '1px solid rgba(16, 185, 129, 0.4)' }}>
              <span style={{ fontSize: 22 }}>🤖</span>
              <div>
                <div style={{ fontSize: 11, fontWeight: 800, color: '#10B981' }}>AI PREDICTION ENGINE</div>
                <div style={{ fontSize: 10, color: '#A7F3D0' }}>REAL-TIME AUTONOMOUS SHIELD</div>
              </div>
            </div>
          </div>

          {/* Map Ground Plane Simulation */}
          <div
            style={{
              position: 'absolute',
              top: 170,
              left: 60,
              right: 60,
              height: 570,
              backgroundColor: 'rgba(10, 15, 29, 0.88)',
              borderRadius: 18,
              border: '2px solid rgba(16, 185, 129, 0.5)',
              boxShadow: '0 12px 40px rgba(16, 185, 129, 0.25)',
              overflow: 'hidden',
              zIndex: 10,
            }}
          >
            {/* Dirt road */}
            <div
              style={{
                position: 'absolute',
                top: '52%',
                left: 0,
                right: 0,
                height: 70,
                transform: 'translateY(-50%)',
                backgroundColor: 'rgba(51, 65, 85, 0.35)',
                borderTop: '2px dashed rgba(148, 163, 184, 0.3)',
                borderBottom: '2px dashed rgba(148, 163, 184, 0.3)',
              }}
            />

            {/* Target Attractant (e.g. Sugarcane Field) */}
            <div
              style={{
                position: 'absolute',
                top: 40,
                right: 60,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                backgroundColor: 'rgba(30, 41, 59, 0.85)',
                border: '1px solid #F59E0B',
                borderRadius: 10,
                padding: '8px 14px',
              }}
            >
              <div style={{ fontSize: 24 }}>🌾</div>
              <div style={{ fontSize: 11, fontWeight: 800, color: '#F59E0B', marginTop: 2 }}>
                RIPENING SUGARCANE
              </div>
              <div style={{ fontSize: 10, color: '#94A3B8' }}>STRONG SCENT ATTRACTANT</div>
            </div>

            {/* Moving Tiger with 3-Tier Dynamic Concentric Geofences */}
            <div
              style={{
                position: 'absolute',
                left: tigerX3,
                top: tigerY3,
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 20,
              }}
            >
              {/* TIER 1: Advisory Buffer (Cyan / Yellow, 360px) */}
              <div
                style={{
                  position: 'absolute',
                  width: 360,
                  height: 360,
                  borderRadius: '50%',
                  border: '2px dashed rgba(56, 189, 248, 0.5)',
                  backgroundColor: 'rgba(56, 189, 248, 0.05)',
                  boxShadow: `0 0 ${15 + pulse * 15}px rgba(56, 189, 248, 0.2)`,
                }}
              />

              {/* TIER 2: Caution Buffer (Orange, 230px) */}
              <div
                style={{
                  position: 'absolute',
                  width: 230,
                  height: 230,
                  borderRadius: '50%',
                  border: '2px solid rgba(245, 158, 11, 0.8)',
                  backgroundColor: 'rgba(245, 158, 11, 0.1)',
                }}
              />

              {/* TIER 3: Danger Zone (Crimson Red, 130px) */}
              <div
                style={{
                  position: 'absolute',
                  width: 130,
                  height: 130,
                  borderRadius: '50%',
                  border: '3px solid #EF4444',
                  backgroundColor: 'rgba(239, 68, 68, 0.25)',
                  boxShadow: `0 0 ${20 + pulse * 20}px rgba(239, 68, 68, 0.6)`,
                }}
              />

              {/* Predictive Trajectory Arrow / Cone pointing forward */}
              <svg
                width="240"
                height="100"
                style={{
                  position: 'absolute',
                  left: 20,
                  top: -50,
                  pointerEvents: 'none',
                  transform: `rotate(${predAngle}deg)`,
                }}
              >
                <polygon
                  points="0,50 140,10 190,50 140,90"
                  fill="rgba(16, 185, 129, 0.15)"
                  stroke="#10B981"
                  strokeWidth="2"
                  strokeDasharray="6 4"
                />
                <line
                  x1="0"
                  y1="50"
                  x2="180"
                  y2="50"
                  stroke="#10B981"
                  strokeWidth="3"
                />
              </svg>

              {/* Tiger Avatar Node */}
              <div
                style={{
                  width: 54,
                  height: 54,
                  borderRadius: '50%',
                  backgroundColor: '#EA580C',
                  border: '3px solid #fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 28,
                  boxShadow: '0 4px 20px rgba(0,0,0,0.8)',
                  zIndex: 25,
                }}
              >
                🐅
              </div>
            </div>

            {/* Detouring Worker with Green Safe Route */}
            <div
              style={{
                position: 'absolute',
                left: 1040,
                top: 260,
                transform: 'translate(-50%, -50%)',
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                zIndex: 25,
              }}
            >
              <div
                style={{
                  width: 46,
                  height: 46,
                  borderRadius: '50%',
                  backgroundColor: '#10B981',
                  border: '2px solid #fff',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 22,
                  boxShadow: '0 0 15px #10B981',
                }}
              >
                🚶
              </div>
              <div
                style={{
                  marginTop: 6,
                  background: 'rgba(16, 185, 129, 0.95)',
                  color: '#fff',
                  padding: '3px 10px',
                  borderRadius: 6,
                  fontSize: 11,
                  fontWeight: 800,
                  whiteSpace: 'nowrap',
                }}
              >
                ✓ SAFE DETOUR IN PROGRESS
              </div>
            </div>

            {/* Bottom Telemetry Legend inside the map container */}
            <div
              style={{
                position: 'absolute',
                bottom: 14,
                left: 20,
                right: 20,
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                backgroundColor: 'rgba(15, 23, 42, 0.92)',
                border: '1px solid rgba(255,255,255,0.1)',
                borderRadius: 8,
                padding: '8px 16px',
              }}
            >
              <div style={{ display: 'flex', gap: 20, alignItems: 'center' }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={{ width: 12, height: 12, borderRadius: '50%', border: '2px dashed #38BDF8', backgroundColor: 'rgba(56,189,248,0.3)' }} />
                  <span style={{ fontSize: 11, color: '#94A3B8' }}>TIER 1 (ADVISORY // 1000m)</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={{ width: 12, height: 12, borderRadius: '50%', border: '2px solid #F59E0B', backgroundColor: 'rgba(245,158,11,0.4)' }} />
                  <span style={{ fontSize: 11, color: '#F59E0B', fontWeight: 700 }}>TIER 2 (CAUTION // 500m)</span>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <div style={{ width: 12, height: 12, borderRadius: '50%', border: '2px solid #EF4444', backgroundColor: 'rgba(239,68,68,0.5)' }} />
                  <span style={{ fontSize: 11, color: '#EF4444', fontWeight: 800 }}>TIER 3 (DANGER // 150m)</span>
                </div>
              </div>
              <div style={{ color: '#10B981', fontSize: 12, fontWeight: 800 }}>
                ● 100% AUTONOMOUS // ZERO DEADLY ENCOUNTERS
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  );
};
