/**
 * Notification sounds using Web Audio API.
 * No audio files needed - sounds are synthesized on demand.
 */

export type NotificationSoundId =
  // Silent
  | 'silent'
  // Simple
  | 'ding'
  | 'chime-up'
  | 'chime-down'
  | 'pop'
  | 'bubble'
  // Playful
  | 'retro'
  | 'coin'
  | 'powerup'
  | 'fanfare'
  | 'laser'
  | 'la-cucaracha'
  // Robots
  | 'robot'
  | 'ufo'
  | 'beepboop'
  | 'dialup'
  | 'r2d2'
  // Musical
  | 'harp'
  | 'music-box'
  | 'celesta'
  | 'synth'
  | 'orchestra'
  // Here Be Dragons
  | 'chaos'
  | 'glitch'
  | 'siren'
  | 'dubstep'
  | 'circus';

export type SoundCategory =
  | 'Silent'
  | 'Simple'
  | 'Playful'
  | 'Robots'
  | 'Musical'
  | 'Here Be Dragons';

export interface NotificationSound {
  id: NotificationSoundId;
  name: string;
  category: SoundCategory;
}

export const notificationSounds: NotificationSound[] = [
  // Silent
  { id: 'silent', name: 'Silent', category: 'Silent' },

  // Simple - clean, professional notification sounds
  { id: 'ding', name: 'Ding', category: 'Simple' },
  { id: 'chime-up', name: 'Rising Chime', category: 'Simple' },
  { id: 'chime-down', name: 'Falling Chime', category: 'Simple' },
  { id: 'pop', name: 'Soft Pop', category: 'Simple' },
  { id: 'bubble', name: 'Bubble', category: 'Simple' },

  // Playful - retro gaming sounds
  { id: 'retro', name: 'Retro Beep', category: 'Playful' },
  { id: 'coin', name: 'Coin Collect', category: 'Playful' },
  { id: 'powerup', name: 'Power Up', category: 'Playful' },
  { id: 'fanfare', name: '8-bit Fanfare', category: 'Playful' },
  { id: 'laser', name: 'Laser Zap', category: 'Playful' },

  // Robots - bleeps, bloops, and digital voices
  { id: 'robot', name: 'Robot Voice', category: 'Robots' },
  { id: 'ufo', name: 'UFO', category: 'Robots' },
  { id: 'beepboop', name: 'Beep Boop', category: 'Robots' },
  { id: 'dialup', name: 'Dial-Up', category: 'Robots' },
  { id: 'r2d2', name: 'R2-D2', category: 'Robots' },

  // Musical - melodies and harmonies
  { id: 'harp', name: 'Harp Flourish', category: 'Musical' },
  { id: 'music-box', name: 'Music Box', category: 'Musical' },
  { id: 'celesta', name: 'Celesta Dream', category: 'Musical' },
  { id: 'synth', name: 'Synth Chord', category: 'Musical' },
  { id: 'orchestra', name: 'Orchestra Hit', category: 'Musical' },
  { id: 'la-cucaracha', name: 'La Cucaracha', category: 'Musical' },

  // Here Be Dragons - absolute madness
  { id: 'chaos', name: 'Chaos', category: 'Here Be Dragons' },
  { id: 'glitch', name: 'Glitch', category: 'Here Be Dragons' },
  { id: 'siren', name: 'Alert Siren', category: 'Here Be Dragons' },
  { id: 'dubstep', name: 'Dubstep Drop', category: 'Here Be Dragons' },
  { id: 'circus', name: 'Circus', category: 'Here Be Dragons' }
];

export const defaultSoundId: NotificationSoundId = 'chime-up';

export const soundCategories: SoundCategory[] = [
  'Silent',
  'Simple',
  'Playful',
  'Robots',
  'Musical',
  'Here Be Dragons'
];

// Lazy-initialized AudioContext (created on first user interaction)
let audioCtx: AudioContext | null = null;

function getContext(): AudioContext {
  if (!audioCtx) {
    audioCtx = new AudioContext();
  }
  // Resume if suspended (browsers suspend until user interaction)
  if (audioCtx.state === 'suspended') {
    audioCtx.resume();
  }
  return audioCtx;
}

/**
 * Play a notification sound by ID.
 * Returns a promise that resolves when the sound finishes.
 */
export function playNotificationSound(soundId: NotificationSoundId): Promise<void> {
  if (soundId === 'silent') {
    return Promise.resolve();
  }

  const player = soundPlayers[soundId];
  if (player) {
    return player();
  }

  console.warn(`Unknown notification sound: ${soundId}`);
  return Promise.resolve();
}

// Sound player functions
const soundPlayers: Record<NotificationSoundId, () => Promise<void>> = {
  silent: () => Promise.resolve(),

  // ============================================================================
  // SIMPLE - Clean, professional notification sounds
  // ============================================================================

  ding: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.frequency.value = 880;
    osc.type = 'sine';

    gain.gain.setValueAtTime(0.3, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.4);

    return delay(400);
  },

  'chime-up': () => {
    const ctx = getContext();

    [659.25, 880].forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.frequency.value = freq;
      osc.type = 'sine';

      const startTime = ctx.currentTime + i * 0.12;
      gain.gain.setValueAtTime(0.25, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.3);

      osc.start(startTime);
      osc.stop(startTime + 0.3);
    });

    return delay(420);
  },

  'chime-down': () => {
    const ctx = getContext();

    [880, 659.25].forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.frequency.value = freq;
      osc.type = 'sine';

      const startTime = ctx.currentTime + i * 0.12;
      gain.gain.setValueAtTime(0.25, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.3);

      osc.start(startTime);
      osc.stop(startTime + 0.3);
    });

    return delay(420);
  },

  pop: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.frequency.setValueAtTime(600, ctx.currentTime);
    osc.frequency.exponentialRampToValueAtTime(200, ctx.currentTime + 0.1);
    osc.type = 'sine';

    gain.gain.setValueAtTime(0.3, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.1);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.1);

    return delay(100);
  },

  bubble: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.frequency.setValueAtTime(400, ctx.currentTime);
    osc.frequency.exponentialRampToValueAtTime(800, ctx.currentTime + 0.1);
    osc.frequency.exponentialRampToValueAtTime(600, ctx.currentTime + 0.15);
    osc.type = 'sine';

    gain.gain.setValueAtTime(0.2, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.2);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.2);

    return delay(200);
  },

  // ============================================================================
  // PLAYFUL - Retro gaming sounds
  // ============================================================================

  retro: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'square';
    osc.frequency.setValueAtTime(440, ctx.currentTime);
    osc.frequency.setValueAtTime(880, ctx.currentTime + 0.1);

    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    gain.gain.setValueAtTime(0.15, ctx.currentTime + 0.1);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.2);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.2);

    return delay(200);
  },

  coin: () => {
    const ctx = getContext();

    // Classic coin sound: first note short, second note sustains longer
    const notes = [
      { freq: 988, dur: 0.15 },
      { freq: 1319, dur: 0.2 }
    ];

    notes.forEach(({ freq, dur }, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.frequency.value = freq;
      osc.type = 'square';

      const startTime = ctx.currentTime + i * 0.08;
      gain.gain.setValueAtTime(0.12, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc.stop(startTime + dur);
    });

    return delay(280);
  },

  powerup: () => {
    const ctx = getContext();

    const notes = [262, 330, 392, 523, 659, 784];
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'square';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + i * 0.05;
      gain.gain.setValueAtTime(0.12, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.1);

      osc.start(startTime);
      osc.stop(startTime + 0.1);
    });

    return delay(400);
  },

  fanfare: () => {
    const ctx = getContext();

    const melody = [
      { freq: 523, time: 0, dur: 0.1 },
      { freq: 659, time: 0.1, dur: 0.1 },
      { freq: 784, time: 0.2, dur: 0.1 },
      { freq: 1047, time: 0.3, dur: 0.25 }
    ];

    melody.forEach(({ freq, time, dur }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'square';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0.12, startTime);
      gain.gain.setValueAtTime(0.12, startTime + dur * 0.8);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc.stop(startTime + dur);
    });

    return delay(550);
  },

  laser: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'sawtooth';
    osc.frequency.setValueAtTime(1500, ctx.currentTime);
    osc.frequency.exponentialRampToValueAtTime(100, ctx.currentTime + 0.15);

    gain.gain.setValueAtTime(0.2, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.15);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.15);

    return delay(150);
  },

  'la-cucaracha': () => {
    const ctx = getContext();

    // Classic "La Cucaracha" horn melody: C-C-C-F-A
    // Three short notes, then two longer ones
    const melody = [
      { freq: 261.63, time: 0, dur: 0.08 }, // C4 - "La"
      { freq: 261.63, time: 0.1, dur: 0.08 }, // C4 - "cu"
      { freq: 261.63, time: 0.2, dur: 0.08 }, // C4 - "ca"
      { freq: 349.23, time: 0.32, dur: 0.15 }, // F4 - "ra"
      { freq: 440, time: 0.5, dur: 0.2 } // A4 - "cha"
    ];

    melody.forEach(({ freq, time, dur }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'square';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0.12, startTime);
      gain.gain.setValueAtTime(0.1, startTime + dur * 0.7);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc.stop(startTime + dur);
    });

    return delay(700);
  },

  // ============================================================================
  // ROBOTS - Bleeps, bloops, and digital voices
  // ============================================================================

  robot: () => {
    const ctx = getContext();

    const notes = [200, 250, 200, 300];
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const osc2 = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      osc2.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'square';
      osc2.type = 'square';
      osc.frequency.value = freq;
      osc2.frequency.value = freq * 1.01;

      const startTime = ctx.currentTime + i * 0.08;
      gain.gain.setValueAtTime(0.1, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.07);

      osc.start(startTime);
      osc2.start(startTime);
      osc.stop(startTime + 0.07);
      osc2.stop(startTime + 0.07);
    });

    return delay(350);
  },

  ufo: () => {
    const ctx = getContext();
    const osc = ctx.createOscillator();
    const lfo = ctx.createOscillator();
    const lfoGain = ctx.createGain();
    const gain = ctx.createGain();

    lfo.connect(lfoGain);
    lfoGain.connect(osc.frequency);

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'sine';
    osc.frequency.value = 600;

    lfo.type = 'sine';
    lfo.frequency.value = 8;
    lfoGain.gain.value = 150;

    gain.gain.setValueAtTime(0.2, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);

    osc.start(ctx.currentTime);
    lfo.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.4);
    lfo.stop(ctx.currentTime + 0.4);

    return delay(400);
  },

  beepboop: () => {
    const ctx = getContext();

    // Classic robot "beep boop" pattern
    const pattern = [
      { freq: 800, dur: 0.08 },
      { freq: 600, dur: 0.12 },
      { freq: 900, dur: 0.06 },
      { freq: 400, dur: 0.15 }
    ];

    let time = 0;
    pattern.forEach(({ freq, dur }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'square';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0.15, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur - 0.01);

      osc.start(startTime);
      osc.stop(startTime + dur);
      time += dur + 0.02;
    });

    return delay(500);
  },

  dialup: () => {
    const ctx = getContext();

    // Simulated dial-up modem handshake sounds
    const osc = ctx.createOscillator();
    const osc2 = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    osc2.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'square';
    osc2.type = 'sawtooth';

    // Frequency modulation to simulate modem tones
    osc.frequency.setValueAtTime(1200, ctx.currentTime);
    osc.frequency.setValueAtTime(2400, ctx.currentTime + 0.1);
    osc.frequency.setValueAtTime(1800, ctx.currentTime + 0.2);
    osc.frequency.setValueAtTime(300, ctx.currentTime + 0.3);

    osc2.frequency.setValueAtTime(2100, ctx.currentTime);
    osc2.frequency.setValueAtTime(1500, ctx.currentTime + 0.15);
    osc2.frequency.setValueAtTime(2700, ctx.currentTime + 0.25);

    gain.gain.setValueAtTime(0.08, ctx.currentTime);
    gain.gain.setValueAtTime(0.1, ctx.currentTime + 0.2);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.4);

    osc.start(ctx.currentTime);
    osc2.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.4);
    osc2.stop(ctx.currentTime + 0.4);

    return delay(400);
  },

  r2d2: () => {
    const ctx = getContext();

    // R2-D2 style excited beeping
    const beeps = [
      { freq: 1800, time: 0, dur: 0.06 },
      { freq: 2200, time: 0.08, dur: 0.04 },
      { freq: 1600, time: 0.14, dur: 0.08 },
      { freq: 2400, time: 0.24, dur: 0.05 },
      { freq: 1400, time: 0.31, dur: 0.1 },
      { freq: 2000, time: 0.43, dur: 0.07 }
    ];

    beeps.forEach(({ freq, time, dur }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'sine';
      // Add slight frequency wobble
      osc.frequency.setValueAtTime(freq, ctx.currentTime + time);
      osc.frequency.linearRampToValueAtTime(freq * 1.1, ctx.currentTime + time + dur / 2);
      osc.frequency.linearRampToValueAtTime(freq * 0.95, ctx.currentTime + time + dur);

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0.2, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc.stop(startTime + dur);
    });

    return delay(550);
  },

  // ============================================================================
  // MUSICAL - Melodies and harmonies
  // ============================================================================

  harp: () => {
    const ctx = getContext();

    const notes = [523, 659, 784, 1047, 1319, 1568, 1319, 1047];
    notes.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const osc2 = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      osc2.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'triangle';
      osc2.type = 'sine';
      osc.frequency.value = freq;
      osc2.frequency.value = freq * 2;

      const startTime = ctx.currentTime + i * 0.04;
      const noteDur = 0.3;

      gain.gain.setValueAtTime(0, startTime);
      gain.gain.linearRampToValueAtTime(0.15, startTime + 0.01);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + noteDur);

      osc.start(startTime);
      osc2.start(startTime);
      osc.stop(startTime + noteDur);
      osc2.stop(startTime + noteDur);
    });

    return delay(620);
  },

  'music-box': () => {
    const ctx = getContext();

    const melody = [
      { freq: 1319, time: 0 },
      { freq: 1175, time: 0.12 },
      { freq: 1319, time: 0.24 },
      { freq: 988, time: 0.36 },
      { freq: 1047, time: 0.48 },
      { freq: 880, time: 0.6 }
    ];

    melody.forEach(({ freq, time }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'sine';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0, startTime);
      gain.gain.linearRampToValueAtTime(0.2, startTime + 0.005);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.25);

      osc.start(startTime);
      osc.stop(startTime + 0.25);

      const harm = ctx.createOscillator();
      const harmGain = ctx.createGain();
      harm.connect(harmGain);
      harmGain.connect(ctx.destination);

      harm.type = 'sine';
      harm.frequency.value = freq * 2;

      harmGain.gain.setValueAtTime(0, startTime);
      harmGain.gain.linearRampToValueAtTime(0.05, startTime + 0.005);
      harmGain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.15);

      harm.start(startTime);
      harm.stop(startTime + 0.15);
    });

    return delay(850);
  },

  celesta: () => {
    const ctx = getContext();

    const melody = [
      { freq: 1047, time: 0, dur: 0.3 },
      { freq: 1175, time: 0.08, dur: 0.25 },
      { freq: 1319, time: 0.16, dur: 0.35 },
      { freq: 1568, time: 0.28, dur: 0.3 },
      { freq: 1760, time: 0.4, dur: 0.25 },
      { freq: 2093, time: 0.52, dur: 0.5 }
    ];

    melody.forEach(({ freq, time, dur }) => {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'sine';
      osc.frequency.value = freq;

      const startTime = ctx.currentTime + time;
      gain.gain.setValueAtTime(0, startTime);
      gain.gain.linearRampToValueAtTime(0.15, startTime + 0.005);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc.stop(startTime + dur);

      [1.5, 2, 3].forEach((mult, j) => {
        const harm = ctx.createOscillator();
        const harmGain = ctx.createGain();

        harm.connect(harmGain);
        harmGain.connect(ctx.destination);

        harm.type = 'sine';
        harm.frequency.value = freq * mult;

        const vol = 0.04 / (j + 1);
        harmGain.gain.setValueAtTime(0, startTime);
        harmGain.gain.linearRampToValueAtTime(vol, startTime + 0.003);
        harmGain.gain.exponentialRampToValueAtTime(0.001, startTime + dur * 0.7);

        harm.start(startTime);
        harm.stop(startTime + dur * 0.7);
      });
    });

    return delay(1000);
  },

  synth: () => {
    const ctx = getContext();

    const chord = [440, 523.25, 659.25, 783.99];

    chord.forEach((freq, i) => {
      const osc = ctx.createOscillator();
      const osc2 = ctx.createOscillator();
      const filter = ctx.createBiquadFilter();
      const gain = ctx.createGain();

      osc.connect(filter);
      osc2.connect(filter);
      filter.connect(gain);
      gain.connect(ctx.destination);

      osc.type = 'sawtooth';
      osc2.type = 'sawtooth';
      osc.frequency.value = freq;
      osc2.frequency.value = freq * 1.005;

      filter.type = 'lowpass';
      filter.frequency.setValueAtTime(3000, ctx.currentTime);
      filter.frequency.exponentialRampToValueAtTime(800, ctx.currentTime + 0.3);
      filter.Q.value = 2;

      const startTime = ctx.currentTime + i * 0.02;
      gain.gain.setValueAtTime(0, startTime);
      gain.gain.linearRampToValueAtTime(0.12, startTime + 0.01);
      gain.gain.setValueAtTime(0.1, startTime + 0.1);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.35);

      osc.start(startTime);
      osc2.start(startTime);
      osc.stop(startTime + 0.35);
      osc2.stop(startTime + 0.35);
    });

    return delay(400);
  },

  orchestra: () => {
    const ctx = getContext();

    const notes: Array<{ freq: number; type: OscillatorType }> = [
      { freq: 130.81, type: 'sawtooth' },
      { freq: 196, type: 'sawtooth' },
      { freq: 261.63, type: 'square' },
      { freq: 329.63, type: 'triangle' },
      { freq: 392, type: 'sine' },
      { freq: 523.25, type: 'sine' }
    ];

    notes.forEach(({ freq, type }) => {
      const osc = ctx.createOscillator();
      const filter = ctx.createBiquadFilter();
      const gain = ctx.createGain();

      osc.connect(filter);
      filter.connect(gain);
      gain.connect(ctx.destination);

      osc.type = type;
      osc.frequency.value = freq;

      filter.type = 'lowpass';
      filter.frequency.setValueAtTime(2000, ctx.currentTime);
      filter.Q.value = 0.5;

      gain.gain.setValueAtTime(0, ctx.currentTime);
      gain.gain.linearRampToValueAtTime(0.08, ctx.currentTime + 0.02);
      gain.gain.setValueAtTime(0.06, ctx.currentTime + 0.1);
      gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.5);

      osc.start(ctx.currentTime);
      osc.stop(ctx.currentTime + 0.5);
    });

    return delay(500);
  },

  // ============================================================================
  // HERE BE DRAGONS - Absolute madness
  // ============================================================================

  chaos: () => {
    const ctx = getContext();

    // Random frequency chaos with multiple oscillators
    for (let i = 0; i < 8; i++) {
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      gain.connect(ctx.destination);

      osc.type = ['sine', 'square', 'sawtooth', 'triangle'][
        Math.floor(Math.random() * 4)
      ] as OscillatorType;

      const startFreq = 200 + Math.random() * 2000;
      const endFreq = 200 + Math.random() * 2000;

      osc.frequency.setValueAtTime(startFreq, ctx.currentTime);
      osc.frequency.exponentialRampToValueAtTime(endFreq, ctx.currentTime + 0.3);

      const startTime = ctx.currentTime + Math.random() * 0.1;
      gain.gain.setValueAtTime(0.08, startTime);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.2 + Math.random() * 0.2);

      osc.start(startTime);
      osc.stop(startTime + 0.4);
    }

    return delay(500);
  },

  glitch: () => {
    const ctx = getContext();

    // Digital glitch - rapid frequency jumping with bit-crushed feel
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'square';

    // Rapid frequency changes to simulate glitching
    const glitchFreqs = [100, 2000, 50, 1500, 800, 3000, 200, 1000, 400, 2500];
    glitchFreqs.forEach((freq, i) => {
      osc.frequency.setValueAtTime(freq, ctx.currentTime + i * 0.03);
    });

    gain.gain.setValueAtTime(0.15, ctx.currentTime);
    gain.gain.setValueAtTime(0.12, ctx.currentTime + 0.15);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.3);

    osc.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.3);

    // Add some noise bursts
    for (let i = 0; i < 3; i++) {
      const bufferSize = ctx.sampleRate * 0.05;
      const buffer = ctx.createBuffer(1, bufferSize, ctx.sampleRate);
      const data = buffer.getChannelData(0);

      for (let j = 0; j < bufferSize; j++) {
        data[j] = Math.random() * 2 - 1;
      }

      const noise = ctx.createBufferSource();
      noise.buffer = buffer;

      const noiseGain = ctx.createGain();
      noise.connect(noiseGain);
      noiseGain.connect(ctx.destination);

      const startTime = ctx.currentTime + i * 0.1 + 0.05;
      noiseGain.gain.setValueAtTime(0.1, startTime);
      noiseGain.gain.exponentialRampToValueAtTime(0.001, startTime + 0.04);

      noise.start(startTime);
      noise.stop(startTime + 0.05);
    }

    return delay(350);
  },

  siren: () => {
    const ctx = getContext();

    // Crazy alert siren with multiple phases
    const osc = ctx.createOscillator();
    const osc2 = ctx.createOscillator();
    const gain = ctx.createGain();

    osc.connect(gain);
    osc2.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'sawtooth';
    osc2.type = 'square';

    // Siren sweep up and down rapidly
    const now = ctx.currentTime;
    for (let i = 0; i < 4; i++) {
      const t = now + i * 0.15;
      osc.frequency.setValueAtTime(400, t);
      osc.frequency.exponentialRampToValueAtTime(1200, t + 0.075);
      osc.frequency.exponentialRampToValueAtTime(400, t + 0.15);
    }

    osc2.frequency.setValueAtTime(100, now);
    osc2.frequency.setValueAtTime(150, now + 0.3);
    osc2.frequency.setValueAtTime(100, now + 0.6);

    gain.gain.setValueAtTime(0.15, now);
    gain.gain.setValueAtTime(0.12, now + 0.3);
    gain.gain.exponentialRampToValueAtTime(0.001, now + 0.6);

    osc.start(now);
    osc2.start(now);
    osc.stop(now + 0.6);
    osc2.stop(now + 0.6);

    return delay(600);
  },

  dubstep: () => {
    const ctx = getContext();

    // Wobble bass drop
    const osc = ctx.createOscillator();
    const lfo = ctx.createOscillator();
    const lfoGain = ctx.createGain();
    const filter = ctx.createBiquadFilter();
    const gain = ctx.createGain();

    lfo.connect(lfoGain);
    lfoGain.connect(filter.frequency);

    osc.connect(filter);
    filter.connect(gain);
    gain.connect(ctx.destination);

    osc.type = 'sawtooth';
    osc.frequency.setValueAtTime(55, ctx.currentTime); // Low bass

    // Wobble LFO
    lfo.type = 'sine';
    lfo.frequency.setValueAtTime(4, ctx.currentTime);
    lfo.frequency.linearRampToValueAtTime(12, ctx.currentTime + 0.3);
    lfo.frequency.linearRampToValueAtTime(6, ctx.currentTime + 0.6);

    lfoGain.gain.value = 800;

    filter.type = 'lowpass';
    filter.frequency.setValueAtTime(500, ctx.currentTime);
    filter.Q.value = 8;

    gain.gain.setValueAtTime(0.25, ctx.currentTime);
    gain.gain.setValueAtTime(0.2, ctx.currentTime + 0.3);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.7);

    osc.start(ctx.currentTime);
    lfo.start(ctx.currentTime);
    osc.stop(ctx.currentTime + 0.7);
    lfo.stop(ctx.currentTime + 0.7);

    // Add a sub-bass thump
    const sub = ctx.createOscillator();
    const subGain = ctx.createGain();
    sub.connect(subGain);
    subGain.connect(ctx.destination);

    sub.type = 'sine';
    sub.frequency.setValueAtTime(60, ctx.currentTime);
    sub.frequency.exponentialRampToValueAtTime(30, ctx.currentTime + 0.2);

    subGain.gain.setValueAtTime(0.3, ctx.currentTime);
    subGain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.25);

    sub.start(ctx.currentTime);
    sub.stop(ctx.currentTime + 0.25);

    return delay(700);
  },

  circus: () => {
    const ctx = getContext();

    // Chaotic circus calliope melody
    const melody = [
      { freq: 523, time: 0 }, // C5
      { freq: 587, time: 0.08 }, // D5
      { freq: 659, time: 0.16 }, // E5
      { freq: 698, time: 0.24 }, // F5
      { freq: 784, time: 0.32 }, // G5
      { freq: 698, time: 0.4 }, // F5
      { freq: 659, time: 0.48 }, // E5
      { freq: 587, time: 0.56 }, // D5
      { freq: 523, time: 0.64 }, // C5
      { freq: 784, time: 0.72 } // G5 (honk!)
    ];

    melody.forEach(({ freq, time }, i) => {
      const osc = ctx.createOscillator();
      const osc2 = ctx.createOscillator();
      const gain = ctx.createGain();

      osc.connect(gain);
      osc2.connect(gain);
      gain.connect(ctx.destination);

      // Slightly detuned for that out-of-tune calliope feel
      osc.type = 'square';
      osc2.type = 'sawtooth';
      osc.frequency.value = freq * (1 + (Math.random() - 0.5) * 0.02);
      osc2.frequency.value = freq * 2 * (1 + (Math.random() - 0.5) * 0.03);

      const startTime = ctx.currentTime + time;
      const dur = i === melody.length - 1 ? 0.2 : 0.07;

      gain.gain.setValueAtTime(0.1, startTime);
      gain.gain.setValueAtTime(0.08, startTime + dur * 0.7);
      gain.gain.exponentialRampToValueAtTime(0.001, startTime + dur);

      osc.start(startTime);
      osc2.start(startTime);
      osc.stop(startTime + dur);
      osc2.stop(startTime + dur);
    });

    return delay(920);
  }
};

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
