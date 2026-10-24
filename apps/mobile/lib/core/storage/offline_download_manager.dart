import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../platform/native_media_bridge.dart';

enum OfflineDownloadStatus {
  queued,
  downloading,
  paused,
  completed,
  expired,
  revoked,
  failed,
}

class OfflineVodRecord {
  final String videoId;
  final String title;
  final String channelName;
  final String resolution; // '1080p', '720p', '480p'
  final int sizeBytes;
  final int downloadedBytes;
  final String encryptedContainerPath; // `.gvpenc` CENC container, never raw `.mp4`
  final String drmKeySetId;
  final DateTime licenseExpiresAt;
  final DateTime? firstPlaybackStartedAt;
  final OfflineDownloadStatus status;

  const OfflineVodRecord({
    required this.videoId,
    required this.title,
    required this.channelName,
    required this.resolution,
    required this.sizeBytes,
    required this.downloadedBytes,
    required this.encryptedContainerPath,
    required this.drmKeySetId,
    required this.licenseExpiresAt,
    this.firstPlaybackStartedAt,
    required this.status,
  });

  double get progress =>
      sizeBytes > 0 ? (downloadedBytes / sizeBytes).clamp(0.0, 1.0) : 0.0;

  bool isLicenseExpiredAt(DateTime nowUtc) {
    if (nowUtc.isAfter(licenseExpiresAt)) {
      return true;
    }
    if (firstPlaybackStartedAt != null) {
      final rentalWindowEnd =
          firstPlaybackStartedAt!.add(const Duration(hours: 48));
      if (nowUtc.isAfter(rentalWindowEnd)) {
        return true;
      }
    }
    return false;
  }

  OfflineVodRecord copyWith({
    int? downloadedBytes,
    DateTime? licenseExpiresAt,
    DateTime? firstPlaybackStartedAt,
    OfflineDownloadStatus? status,
  }) {
    return OfflineVodRecord(
      videoId: videoId,
      title: title,
      channelName: channelName,
      resolution: resolution,
      sizeBytes: sizeBytes,
      downloadedBytes: downloadedBytes ?? this.downloadedBytes,
      encryptedContainerPath: encryptedContainerPath,
      drmKeySetId: drmKeySetId,
      licenseExpiresAt: licenseExpiresAt ?? this.licenseExpiresAt,
      firstPlaybackStartedAt:
          firstPlaybackStartedAt ?? this.firstPlaybackStartedAt,
      status: status ?? this.status,
    );
  }
}

class OfflineAccessDeniedException implements Exception {
  final String reason;
  const OfflineAccessDeniedException(this.reason);

  @override
  String toString() => 'OfflineAccessDeniedException: $reason';
}

class OfflineDownloadState {
  final bool hasActivePremiumSubscription;
  final int maxQuotaBytes;
  final List<OfflineVodRecord> records;

  const OfflineDownloadState({
    this.hasActivePremiumSubscription = true,
    this.maxQuotaBytes = 10 * 1024 * 1024 * 1024, // 10 GiB default vault quota
    this.records = const [],
  });

  int get usedBytes =>
      records.fold<int>(0, (sum, item) => sum + item.downloadedBytes);

  int get remainingQuotaBytes =>
      (maxQuotaBytes - usedBytes).clamp(0, maxQuotaBytes);

  OfflineDownloadState copyWith({
    bool? hasActivePremiumSubscription,
    int? maxQuotaBytes,
    List<OfflineVodRecord>? records,
  }) {
    return OfflineDownloadState(
      hasActivePremiumSubscription:
          hasActivePremiumSubscription ?? this.hasActivePremiumSubscription,
      maxQuotaBytes: maxQuotaBytes ?? this.maxQuotaBytes,
      records: records ?? this.records,
    );
  }
}

final offlineDownloadManagerProvider =
    StateNotifierProvider<OfflineDownloadManager, OfflineDownloadState>((ref) {
  final bridge = ref.watch(nativeMediaBridgeProvider);
  return OfflineDownloadManager(mediaBridge: bridge)..seedInitialDownloads();
});

class OfflineDownloadManager extends StateNotifier<OfflineDownloadState> {
  final NativeMediaBridge mediaBridge;

  OfflineDownloadManager({
    required this.mediaBridge,
    OfflineDownloadState? initialState,
  }) : super(initialState ?? const OfflineDownloadState());

  void seedInitialDownloads() {
    if (state.records.isNotEmpty) return;
    final now = DateTime.now().toUtc();
    state = state.copyWith(
      records: [
        OfflineVodRecord(
          videoId: 'vid_offline_av1_01',
          title: 'Designing Low-Latency Global Video Mesh (1080p HDR)',
          channelName: 'GVP Systems Architecture',
          resolution: '1080p',
          sizeBytes: 420 * 1024 * 1024,
          downloadedBytes: 420 * 1024 * 1024,
          encryptedContainerPath:
              '/app_sandbox/vault/vid_offline_av1_01.gvpenc',
          drmKeySetId: 'ks_hw_widevine_001',
          licenseExpiresAt: now.add(const Duration(days: 24)),
          status: OfflineDownloadStatus.completed,
        ),
        OfflineVodRecord(
          videoId: 'vid_offline_doc_02',
          title: 'Deep Ocean Bioluminescence in Dolby Vision',
          channelName: 'Cinema4K HDR Lab',
          resolution: '720p',
          sizeBytes: 280 * 1024 * 1024,
          downloadedBytes: 280 * 1024 * 1024,
          encryptedContainerPath:
              '/app_sandbox/vault/vid_offline_doc_02.gvpenc',
          drmKeySetId: 'ks_hw_widevine_002',
          licenseExpiresAt: now.add(const Duration(days: 14)),
          status: OfflineDownloadStatus.completed,
        ),
      ],
    );
  }

  /// Enqueues an encrypted CENC VOD download. Rejects non-Premium users, quota overflows,
  /// and raw unprotected `.mp4` target containers.
  Future<OfflineVodRecord> enqueueDownload({
    required String videoId,
    required String title,
    required String channelName,
    required String resolution,
    required int estimatedSizeBytes,
    required String manifestUrl,
    required String licenseServerUrl,
    String? customContainerFilename,
  }) async {
    if (!state.hasActivePremiumSubscription) {
      throw const OfflineAccessDeniedException(
        'Active GVP Premium subscription is required for offline downloads.',
      );
    }

    final containerName = customContainerFilename ?? '$videoId.gvpenc';
    if (containerName.toLowerCase().endsWith('.mp4')) {
      throw const OfflineAccessDeniedException(
        'Unprotected raw .mp4 exports are prohibited; offline assets must use hardware-encrypted CENC (.gvpenc).',
      );
    }

    if (estimatedSizeBytes > state.remainingQuotaBytes) {
      throw const OfflineAccessDeniedException(
        'Insufficient offline storage quota remaining.',
      );
    }

    final license = await mediaBridge.acquireOfflineDrmLicense(
      videoId: videoId,
      manifestUrl: manifestUrl,
      licenseServerUrl: licenseServerUrl,
    );

    final record = OfflineVodRecord(
      videoId: videoId,
      title: title,
      channelName: channelName,
      resolution: resolution,
      sizeBytes: estimatedSizeBytes,
      downloadedBytes: estimatedSizeBytes,
      encryptedContainerPath: '/app_sandbox/vault/$containerName',
      drmKeySetId: license.keySetId,
      licenseExpiresAt: license.expiresAt,
      status: OfflineDownloadStatus.completed,
    );

    state = state.copyWith(
      records: [
        ...state.records.where((r) => r.videoId != videoId),
        record,
      ],
    );

    return record;
  }

  /// Validates subscription status and DRM license window before starting offline playback.
  Future<OfflineVodRecord> validateAndStartOfflinePlayback(
    String videoId, {
    DateTime? currentTimeUtc,
  }) async {
    final now = currentTimeUtc ?? DateTime.now().toUtc();

    if (!state.hasActivePremiumSubscription) {
      await revokeAllDownloadsOnSubscriptionExpiry();
      throw const OfflineAccessDeniedException(
        'Offline DRM licenses revoked because GVP Premium subscription expired.',
      );
    }

    final index = state.records.indexWhere((r) => r.videoId == videoId);
    if (index == -1) {
      throw const OfflineAccessDeniedException('Offline asset not found.');
    }

    final record = state.records[index];
    if (record.status == OfflineDownloadStatus.revoked) {
      throw const OfflineAccessDeniedException(
        'Offline DRM license has been revoked.',
      );
    }

    if (record.isLicenseExpiredAt(now)) {
      final updated = record.copyWith(status: OfflineDownloadStatus.expired);
      final nextList = [...state.records];
      nextList[index] = updated;
      state = state.copyWith(records: nextList);
      throw const OfflineAccessDeniedException(
        'Offline DRM license has expired. Reconnect to the internet to renew.',
      );
    }

    final hardwareValid =
        await mediaBridge.verifyOfflineDrmLicense(record.drmKeySetId);
    if (!hardwareValid) {
      throw const OfflineAccessDeniedException(
        'Hardware CDM rejected offline keySetId.',
      );
    }

    final startedRecord = record.copyWith(
      firstPlaybackStartedAt: record.firstPlaybackStartedAt ?? now,
    );
    final nextList = [...state.records];
    nextList[index] = startedRecord;
    state = state.copyWith(records: nextList);
    return startedRecord;
  }

  /// Revokes all persistent DRM keys immediately when a Premium subscription lapses.
  Future<void> revokeAllDownloadsOnSubscriptionExpiry() async {
    for (final record in state.records) {
      await mediaBridge.releaseOfflineDrmLicense(record.drmKeySetId);
    }
    state = state.copyWith(
      hasActivePremiumSubscription: false,
      records: state.records
          .map((r) => r.copyWith(status: OfflineDownloadStatus.revoked))
          .toList(),
    );
  }

  Future<void> renewLicense(String videoId) async {
    if (!state.hasActivePremiumSubscription) {
      throw const OfflineAccessDeniedException(
        'Cannot renew DRM license without an active GVP Premium subscription.',
      );
    }
    final index = state.records.indexWhere((r) => r.videoId == videoId);
    if (index == -1) return;

    final current = state.records[index];
    final renewed = await mediaBridge.acquireOfflineDrmLicense(
      videoId: videoId,
      manifestUrl: 'https://cdn.gvp.video/hls/$videoId/master.m3u8',
      licenseServerUrl: 'https://drm.gvp.video/v1/license',
    );
    final updated = current.copyWith(
      licenseExpiresAt: renewed.expiresAt,
      status: OfflineDownloadStatus.completed,
    );
    final nextList = [...state.records];
    nextList[index] = updated;
    state = state.copyWith(records: nextList);
  }

  Future<void> removeDownload(String videoId) async {
    final target =
        state.records.where((r) => r.videoId == videoId).firstOrNull;
    if (target != null) {
      await mediaBridge.releaseOfflineDrmLicense(target.drmKeySetId);
    }
    state = state.copyWith(
      records: state.records.where((r) => r.videoId != videoId).toList(),
    );
  }
}
