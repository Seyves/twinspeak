export interface WhisperSegment {
  id: number;
  seek: number;
  start: number;
  end: number;
  text: string;
  avg_logprob: number;
  compression_ratio: number;
  no_speech_prob: number;
  completed: boolean;
}

export interface WhisperMessage {
  message?: string;
  segments?: WhisperSegment[];
  error?: string;
}

export type TranscriptionStatus =
  | 'idle'
  | 'connecting'
  | 'ready'
  | 'recording'
  | 'stopping'
  | 'error';

export interface TranscriptLine {
  id: number;
  text: string;
  completed: boolean;
}
