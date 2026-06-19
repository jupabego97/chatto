import type { EventEnvelope } from '$lib/eventBus.svelte';
import type { ExpiringAssetUrl } from '$lib/attachments/attachmentUrls';
import { earliestAssetUrlRefreshAt } from '$lib/attachments/attachmentUrls';

export const ROOM_FILES_PAGE_SIZE = 50;

export type RoomFileItem = {
  messageEventId: string;
  threadRootEventId?: string | null;
  createdAt: string;
  attachment: {
    id: string;
    filename: string;
    contentType: string;
    width?: number | null;
    height?: number | null;
    assetUrl: ExpiringAssetUrl;
    thumbnailAssetUrl?: ExpiringAssetUrl | null;
    videoProcessing?: {
      thumbnailAssetUrl?: ExpiringAssetUrl | null;
    } | null;
  };
};

function isImageAttachment(contentType: string): boolean {
  return contentType.startsWith('image/');
}

function isVideoAttachment(contentType: string): boolean {
  return contentType.startsWith('video/');
}

function attachmentAssetUrls(item: RoomFileItem) {
  return [
    item.attachment.assetUrl,
    item.attachment.thumbnailAssetUrl,
    item.attachment.videoProcessing?.thumbnailAssetUrl
  ];
}

export class RoomFilesStore {
  items = $state.raw<RoomFileItem[]>([]);
  totalCount = $state(0);
  hasMore = $state(false);
  isInitialLoading = $state(false);
  isLoadingMore = $state(false);
  isUnsupported = $state(true);

  #roomId = '';

  constructor(_connection?: unknown) {}

  setRoom(roomId: string): void {
    if (this.#roomId === roomId) return;
    this.#roomId = roomId;
    this.items = [];
    this.totalCount = 0;
    this.hasMore = false;
    this.isInitialLoading = false;
    this.isLoadingMore = false;
    this.isUnsupported = true;
  }

  async loadInitial(): Promise<void> {
    this.isInitialLoading = false;
  }

  async loadMore(): Promise<void> {
    this.isLoadingMore = false;
  }

  async refresh(): Promise<void> {
    this.items = [];
    this.totalCount = 0;
    this.hasMore = false;
    this.isInitialLoading = false;
  }

  ingestServerEvent(_serverEvent: EventEnvelope): void {}

  assetUrlFor(item: RoomFileItem): ExpiringAssetUrl {
    return item.attachment.assetUrl;
  }

  thumbnailAssetUrlFor(item: RoomFileItem): ExpiringAssetUrl | null {
    const contentType = item.attachment.contentType;
    if (isVideoAttachment(contentType)) {
      return item.attachment.videoProcessing?.thumbnailAssetUrl ?? null;
    }
    if (!isImageAttachment(contentType)) return null;
    return item.attachment.thumbnailAssetUrl ?? item.attachment.videoProcessing?.thumbnailAssetUrl ?? null;
  }

  get nextAssetUrlRefreshAt(): number | null {
    return earliestAssetUrlRefreshAt(this.items.flatMap((item) => attachmentAssetUrls(item)));
  }

  hasRefreshableStaleUrl(): boolean {
    return false;
  }

  async refreshStaleUrls(): Promise<void> {}

  async refreshUrlsForItem(_item: RoomFileItem): Promise<void> {}
}
