import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:gvp_mobile/app/app.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets(
    'End-to-End Mobile Flow: Home -> Watch -> MiniPlayer -> Shorts -> Upload -> Offline Playback',
    (tester) async {
      await tester.pumpWidget(const ProviderScope(child: GvpMobileApp()));
      await tester.pumpAndSettle();

      // 1. Home Screen: verify video feed and tap first VOD card
      expect(find.byKey(const Key('home_feed_list')), findsOneWidget);
      final firstVideoCard = find.byKey(const Key('video_card_vid_home_1'));
      expect(firstVideoCard, findsOneWidget);
      await tester.tap(firstVideoCard);
      await tester.pumpAndSettle();

      // 2. Watch Screen: verify player controls, download for offline, minimize to mini-player
      expect(find.byKey(const Key('watch_pip_btn')), findsOneWidget);
      expect(find.byKey(const Key('watch_airplay_cast_btn')), findsOneWidget);

      await tester.tap(find.byKey(const Key('watch_offline_download_btn')));
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('watch_minimize_miniplayer_btn')));
      await tester.pumpAndSettle();

      // 3. Navigate to Shorts Tab and swipe vertically
      await tester.tap(find.text('Shorts'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('shorts_page_view')), findsOneWidget);

      await tester.drag(
        find.byKey(const Key('shorts_page_view')),
        const Offset(0, -500),
      );
      await tester.pumpAndSettle();

      // 4. Navigate to Upload Tab and trigger R2 multipart upload
      await tester.tap(find.text('Upload'));
      await tester.pumpAndSettle();
      expect(find.byKey(const Key('upload_title_field')), findsOneWidget);

      final publishBtn = find.byKey(const Key('upload_start_btn'));
      await tester.ensureVisible(publishBtn);
      await tester.tap(publishBtn);
      await tester.pumpAndSettle();
      expect(find.textContaining('COMPLETED'), findsOneWidget);

      // 5. Navigate to Library -> Offline Tab and launch encrypted offline playback
      await tester.tap(find.text('Library'));
      await tester.pumpAndSettle();

      await tester.tap(find.byKey(const Key('library_offline_tab')));
      await tester.pumpAndSettle();

      final offlineItem = find.byKey(const Key('offline_item_vid_home_1'));
      expect(offlineItem, findsOneWidget);
      await tester.tap(offlineItem);
      await tester.pumpAndSettle();

      // Verify offline video opened in Watch player
      expect(find.byKey(const Key('watch_pip_btn')), findsOneWidget);
    },
  );
}
