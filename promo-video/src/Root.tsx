import React from 'react';
import { Composition } from 'remotion';
import { PromoVideo } from './Video';
import timelineData from '../public/audio/timeline.json';

const FPS = 30;

// Calculate total video duration in frames (exact ~239.2s, 3m59s)
const totalDurationSec = timelineData.reduce((acc, item) => acc + item.duration, 0) + 2.0;
const totalFrames = Math.ceil(totalDurationSec * FPS);

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id="CoExistShieldPromo"
        component={PromoVideo as any}
        durationInFrames={totalFrames}
        fps={FPS}
        width={1920}
        height={1080}
        defaultProps={{
          timeline: timelineData,
        }}
      />
    </>
  );
};
