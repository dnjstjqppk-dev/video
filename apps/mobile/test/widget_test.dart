import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gvp_mobile/app/app.dart';
import 'package:gvp_mobile/core/auth/oidc_auth_service.dart';
import 'package:gvp_mobile/core/network/api_client.dart';
import 'package:gvp_mobile/core/platform/native_media_bridge.dart';
import 'package:gvp_mobile/core/storage/offline_download_manager.dart';
import 'package:gvp_mobile/features/shorts/presentation/shorts_screen.dart';

void main() {
  group('API Client & OIDC PKCE Interceptor Tests', () {
    test('attaches Bearer token and X-Request-ID header and refreshes expired token', () async {
      final store = InMemorySecureTokenStore();
      final authService = OidcAuthService(store: store);

      final pkce = await authService.createPkceChallenge();
      expect(pkce.codeVerifier.length, greaterThanOrEqualTo(43));
      expect(pkce.codeChallenge, isNotEmpty);

      await authService.saveTokens(
        OidcTokenSet(
          accessToken: 'initial_access_token',
          refreshToken: 'valid_refresh_token',
          idToken: 'initial_id_token',
          expiresAt: DateTime.now().toUtc().add(const Duration(minutes: 10)),
        ),
      );

      final dio = Dio(BaseOptions(baseUrl: 'https://api.gvp.video/v1'));
      final apiClient = ApiClient(authService: authService, customDio: dio);

      RequestOptions? capturedOptions;
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            capturedOptions = options;
            handler.resolve(
              Response(
                requestOptions: options,
                statusCode: 200,
                data: {'status': 'ok'},
              ),
            );
          },
        ),
      );

      final response = await apiClient.get<Map<String, dynamic>>('/videos/feed');
      expect(response.statusCode, 200);
      expect(capturedOptions?.headers['Authorization'], 'Bearer initial_access_token');
      expect(capturedOptions?.headers['X-Request-ID'], startsWith('req_'));
    });
  });

  group('Shorts Controller Pool Memory Recycling Tests', () {
    test('retains at most 3 controllers (N-1, N, N+1) and disposes off-screen controllers', () {
      final pool = ShortsControllerPool(totalItems: 10);

      pool.onPageChanged(0);
      expect(pool.activeCount, 2); // 0 and 1
      expect(pool.activeControllers[0]!.isPlaying, isTrue);
      expect(pool.activeControllers[1]!.isPlaying, isFalse);

      pool.onPageChanged(1);
      expect(pool.activeCount, 3); // 0, 1, 2
      expect(pool.activeControllers[1]!.isPlaying, isTrue);

      pool.onPageChanged(2);
      expect(pool.activeCount, 3); // 1, 2, 3
      expect(pool.activeControllers.containsKey(0), isFalse);
      expect(pool.disposedHistory, contains(0));

      pool.onPageChanged(5);
      expect(pool.activeCount, 3); // 4, 5, 6
      expect(pool.activeControllers.keys.toSet(), {4, 5, 6});
    });
  });

  group('Offline DRM License Expiry & Protection Enforcement Tests', () {
    test('blocks raw .mp4 exports, enforces 30d/48h expiry, and revokes on subscription lapse', () async {
      final bridge = NativeMediaBridge();
      final manager = OfflineDownloadManager(mediaBridge: bridge);

      // Reject unprotected .mp4 export attempt
      expect(
        () => manager.enqueueDownload(
          videoId: 'vid_raw_mp4',
          title: 'Raw MP4 Test',
          channelName: 'Tester',
          resolution: '1080p',
          estimatedSizeBytes: 50 * 1024 * 1024,
          manifestUrl: 'https://cdn.gvp.video/hls/vid_raw_mp4/master.m3u8',
          licenseServerUrl: 'https://drm.gvp.video/v1/license',
          customContainerFilename: 'unprotected_export.mp4',
        ),
        throwsA(isA<OfflineAccessDeniedException>()),
      );

      // Valid encrypted download
      final saved = await manager.enqueueDownload(
        videoId: 'vid_drm_test',
        title: 'Encrypted CENC Test',
        channelName: 'GVP Security',
        resolution: '1080p',
        estimatedSizeBytes: 120 * 1024 * 1024,
        manifestUrl: 'https://cdn.gvp.video/hls/vid_drm_test/master.m3u8',
        licenseServerUrl: 'https://drm.gvp.video/v1/license',
      );
      expect(saved.encryptedContainerPath, endsWith('.gvpenc'));

      final startUtc = DateTime.now().toUtc();
      await manager.validateAndStartOfflinePlayback(
        'vid_drm_test',
        currentTimeUtc: startUtc,
      );

      // 49 hours after first playback -> 48h rental window expired
      expect(
        () => manager.validateAndStartOfflinePlayback(
          'vid_drm_test',
          currentTimeUtc: startUtc.add(const Duration(hours: 49)),
        ),
        throwsA(isA<OfflineAccessDeniedException>()),
      );

      // Revoke all on subscription lapse
      await manager.revokeAllDownloadsOnSubscriptionExpiry();
      expect(manager.state.hasActivePremiumSubscription, isFalse);
      expect(
        manager.state.records.every(
          (r) => r.status == OfflineDownloadStatus.revoked,
        ),
        isTrue,
      );
    });
  });

  testWidgets('GVP Mobile App renders Home feed and bottom navigation bar', (tester) async {
    await tester.pumpWidget(const ProviderScope(child: GvpMobileApp()));
    await tester.pumpAndSettle();

    expect(find.text('GVP'), findsOneWidget);
    expect(find.text('Home'), findsOneWidget);
    expect(find.text('Shorts'), findsOneWidget);
    expect(find.text('Upload'), findsOneWidget);
    expect(find.text('Subscriptions'), findsOneWidget);
    expect(find.text('Library'), findsOneWidget);
  });
}
