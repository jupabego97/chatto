export type MediaViewerImageItem = {
  kind: 'image';
  id?: string;
  src: string;
  alt?: string;
  filename?: string;
  openUrl?: string;
};

export type MediaViewerVideoItem = {
  kind: 'video';
  id?: string;
  src: string;
  poster?: string | null;
  filename?: string;
  autoLoop?: boolean;
  width?: number | null;
  height?: number | null;
  startTime?: number;
  openUrl?: string;
  source:
    | {
        kind: 'asset';
      }
    | {
        kind: 'variant';
        quality: string;
      };
};

export type MediaViewerItem = MediaViewerImageItem | MediaViewerVideoItem;
