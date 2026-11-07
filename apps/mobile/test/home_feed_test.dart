import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gvp_mobile/core/auth/oidc_auth_service.dart';
import 'package:gvp_mobile/core/network/api_client.dart';
import 'package:gvp_mobile/features/home/presentation/home_screen.dart';

Map<String, dynamic> feedPage(List<Map<String, dynamic>> items, int total) => {
      'items': items,
      'total': total,
      'page': 1,
      'limit': 20,
    };

Map<String, dynamic> videoJson(String id, {String videoType = 'standard'}) => {
      'id': id,
      'title': 'Test video $id',
      'channel_name': 'Test Channel',
      'category': 'Engineering',
      'view_count': 1500,
      'duration_seconds': 754,
      'video_type': videoType,
      'thumbnail_url': 'https://cdn.example.com/t/$id.jpg',
      'published_at': '2026-11-01T10:00:00Z',
    };

ApiClient stubApiClient(Map<String, dynamic> Function(RequestOptions) handler) {
  final dio = Dio();
  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, interceptorHandler) {
        try {
          final data = handler(options);
          interceptorHandler.resolve(
            Response(requestOptions: options, statusCode: 200, data: data),
          );
        } on DioException catch (e) {
          interceptorHandler.reject(e);
        }
      },
    ),
  );
  final store = InMemorySecureTokenStore();
  return ApiClient(
    authService: OidcAuthService(store: store),
    customDio: dio,
  );
}

void main() {
  group('HomeFeedNotifier real API wiring', () {
    test('loads and parses feed items', () async {
      final api = stubApiClient(
        (_) => feedPage([videoJson('vid_1'), videoJson('vid_2', videoType: 'live')], 2),
      );
      final container = ProviderContainer(
        overrides: [apiClientProvider.overrideWithValue(api)],
      );
      addTearDown(container.dispose);

      await container.read(homeFeedProvider.notifier).loadInitial();
      final state = container.read(homeFeedProvider);
      expect(state.isLoading, isFalse);
      expect(state.errorMessage, isNull);
      expect(state.items, hasLength(2));
      expect(state.items[0].title, 'Test video vid_1');
      expect(state.items[0].viewsText, '1.5K views');
      expect(state.items[1].isLive, isTrue);
      expect(state.items[1].durationLabel, 'LIVE');
    });

    test('surfaces API errors with retry intact', () async {
      var failMode = true;
      final api = stubApiClient((options) {
        if (failMode) {
          throw DioException(
            requestOptions: options,
            response: Response(
              requestOptions: options,
              statusCode: 500,
              data: {
                'error': {'code': 'INTERNAL', 'message': 'boom', 'request_id': 'r1'}
              },
            ),
          );
        }
        return feedPage([videoJson('vid_9')], 1);
      });
      final container = ProviderContainer(
        overrides: [apiClientProvider.overrideWithValue(api)],
      );
      addTearDown(container.dispose);

      await container.read(homeFeedProvider.notifier).loadInitial();
      expect(container.read(homeFeedProvider).errorMessage, isNotNull);
      expect(container.read(homeFeedProvider).items, isEmpty);

      failMode = false;
      await container.read(homeFeedProvider.notifier).refreshFeed();
      final state = container.read(homeFeedProvider);
      expect(state.errorMessage, isNull);
      expect(state.items, hasLength(1));
    });

    test('paginates until total reached', () async {
      final api = stubApiClient((options) {
        final page = (options.queryParameters['page'] as num).toInt();
        if (page == 1) return feedPage([videoJson('vid_1')], 2);
        return feedPage([videoJson('vid_2')], 2);
      });
      final container = ProviderContainer(
        overrides: [apiClientProvider.overrideWithValue(api)],
      );
      addTearDown(container.dispose);

      await container.read(homeFeedProvider.notifier).loadInitial();
      expect(container.read(homeFeedProvider).hasMore, isTrue);
      await container.read(homeFeedProvider.notifier).loadNextPage();
      final state = container.read(homeFeedProvider);
      expect(state.items, hasLength(2));
      expect(state.hasMore, isFalse);
    });
  });
}
