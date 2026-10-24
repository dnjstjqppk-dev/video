import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../features/home/presentation/home_screen.dart';
import '../features/library/presentation/library_screen.dart';
import '../features/live/presentation/live_screen.dart';
import '../features/shorts/presentation/shorts_screen.dart';
import '../features/upload/presentation/upload_screen.dart';
import '../features/watch/presentation/watch_screen.dart';

final miniPlayerVideoProvider = StateProvider<MiniPlayerState?>((ref) => null);

class MiniPlayerState {
  final String videoId;
  final String title;
  final String channelName;
  final bool isPlaying;
  final double progress;

  const MiniPlayerState({
    required this.videoId,
    required this.title,
    required this.channelName,
    this.isPlaying = true,
    this.progress = 0.0,
  });

  MiniPlayerState copyWith({
    bool? isPlaying,
    double? progress,
  }) {
    return MiniPlayerState(
      videoId: videoId,
      title: title,
      channelName: channelName,
      isPlaying: isPlaying ?? this.isPlaying,
      progress: progress ?? this.progress,
    );
  }
}

final appRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    routes: [
      ShellRoute(
        builder: (context, state, child) {
          return MainShellScaffold(
            currentLocation: state.uri.path,
            child: child,
          );
        },
        routes: [
          GoRoute(
            path: '/',
            name: 'home',
            builder: (context, state) => const HomeScreen(),
          ),
          GoRoute(
            path: '/shorts',
            name: 'shorts',
            builder: (context, state) => const ShortsScreen(),
          ),
          GoRoute(
            path: '/upload',
            name: 'upload',
            builder: (context, state) => const UploadScreen(),
          ),
          GoRoute(
            path: '/subscriptions',
            name: 'subscriptions',
            builder: (context, state) => const SubscriptionsScreen(),
          ),
          GoRoute(
            path: '/library',
            name: 'library',
            builder: (context, state) => const LibraryScreen(),
          ),
        ],
      ),
      GoRoute(
        path: '/watch/:videoId',
        name: 'watch',
        builder: (context, state) {
          final videoId = state.pathParameters['videoId'] ?? 'vid_default_01';
          final title = state.uri.queryParameters['title'] ??
              'Global Architecture Deep Dive: 4K 60fps Adaptive Streaming';
          final channel =
              state.uri.queryParameters['channel'] ?? 'GVP Engineering';
          return WatchScreen(
            videoId: videoId,
            initialTitle: title,
            initialChannel: channel,
          );
        },
      ),
      GoRoute(
        path: '/live/:streamId',
        name: 'live',
        builder: (context, state) {
          final streamId = state.pathParameters['streamId'] ?? 'live_main_01';
          final isBroadcaster =
              state.uri.queryParameters['mode'] == 'broadcaster';
          return LiveScreen(
            streamId: streamId,
            initialBroadcasterMode: isBroadcaster,
          );
        },
      ),
      GoRoute(
        path: '/channel/:handle',
        name: 'channel',
        builder: (context, state) => ChannelScreen(
          handle: state.pathParameters['handle'] ?? '@gvp_creators',
        ),
      ),
      GoRoute(
        path: '/search',
        name: 'search',
        builder: (context, state) => const SearchScreen(),
      ),
      GoRoute(
        path: '/notifications',
        name: 'notifications',
        builder: (context, state) => const NotificationsScreen(),
      ),
      GoRoute(
        path: '/downloads',
        name: 'downloads',
        builder: (context, state) => const LibraryScreen(initialTabIndex: 3),
      ),
      GoRoute(
        path: '/premium',
        name: 'premium',
        builder: (context, state) => const PremiumScreen(),
      ),
      GoRoute(
        path: '/settings',
        name: 'settings',
        builder: (context, state) => const SettingsScreen(),
      ),
      GoRoute(
        path: '/onboarding',
        name: 'onboarding',
        builder: (context, state) => const OnboardingScreen(),
      ),
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
    ],
  );
});

class MainShellScaffold extends ConsumerWidget {
  final String currentLocation;
  final Widget child;

  const MainShellScaffold({
    super.key,
    required this.currentLocation,
    required this.child,
  });

  int _locationToIndex(String path) {
    if (path.startsWith('/shorts')) return 1;
    if (path.startsWith('/upload')) return 2;
    if (path.startsWith('/subscriptions')) return 3;
    if (path.startsWith('/library')) return 4;
    return 0;
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selectedIndex = _locationToIndex(currentLocation);
    final miniPlayer = ref.watch(miniPlayerVideoProvider);

    return Scaffold(
      body: Stack(
        children: [
          child,
          if (miniPlayer != null && !currentLocation.startsWith('/shorts'))
            Positioned(
              left: 8,
              right: 8,
              bottom: 8,
              child: _MiniPlayerDock(state: miniPlayer),
            ),
        ],
      ),
      bottomNavigationBar: NavigationBar(
        selectedIndex: selectedIndex,
        onDestinationSelected: (index) {
          switch (index) {
            case 0:
              context.go('/');
            case 1:
              context.go('/shorts');
            case 2:
              context.go('/upload');
            case 3:
              context.go('/subscriptions');
            case 4:
              context.go('/library');
          }
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home_rounded),
            label: 'Home',
          ),
          NavigationDestination(
            icon: Icon(Icons.bolt_outlined),
            selectedIcon: Icon(Icons.bolt_rounded),
            label: 'Shorts',
          ),
          NavigationDestination(
            icon: Icon(Icons.add_circle_outline_rounded, size: 30),
            selectedIcon: Icon(Icons.add_circle_rounded, size: 30),
            label: 'Upload',
          ),
          NavigationDestination(
            icon: Icon(Icons.subscriptions_outlined),
            selectedIcon: Icon(Icons.subscriptions_rounded),
            label: 'Subscriptions',
          ),
          NavigationDestination(
            icon: Icon(Icons.video_library_outlined),
            selectedIcon: Icon(Icons.video_library_rounded),
            label: 'Library',
          ),
        ],
      ),
    );
  }
}

class _MiniPlayerDock extends ConsumerWidget {
  final MiniPlayerState state;

  const _MiniPlayerDock({required this.state});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Material(
      color: const Color(0xFF1C1C28),
      elevation: 10,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () {
          context.push(
            '/watch/${state.videoId}?title=${Uri.encodeComponent(state.title)}&channel=${Uri.encodeComponent(state.channelName)}',
          );
        },
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
              child: Row(
                children: [
                  Container(
                    width: 84,
                    height: 48,
                    decoration: BoxDecoration(
                      color: const Color(0xFFE11D48).withValues(alpha: 0.25),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(
                      Icons.play_circle_fill_rounded,
                      color: Colors.white,
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          state.title,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(
                            fontWeight: FontWeight.w600,
                            fontSize: 13,
                          ),
                        ),
                        const SizedBox(height: 2),
                        Text(
                          state.channelName,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: TextStyle(
                            color: Colors.grey.shade400,
                            fontSize: 11,
                          ),
                        ),
                      ],
                    ),
                  ),
                  IconButton(
                    icon: Icon(
                      state.isPlaying
                          ? Icons.pause_rounded
                          : Icons.play_arrow_rounded,
                    ),
                    onPressed: () {
                      ref.read(miniPlayerVideoProvider.notifier).state =
                          state.copyWith(isPlaying: !state.isPlaying);
                    },
                  ),
                  IconButton(
                    icon: const Icon(Icons.close_rounded),
                    onPressed: () {
                      ref.read(miniPlayerVideoProvider.notifier).state = null;
                    },
                  ),
                ],
              ),
            ),
            LinearProgressIndicator(
              value: state.progress.clamp(0.0, 1.0),
              minHeight: 2,
              backgroundColor: Colors.white12,
              valueColor:
                  const AlwaysStoppedAnimation<Color>(Color(0xFFE11D48)),
            ),
          ],
        ),
      ),
    );
  }
}

class SubscriptionsScreen extends StatelessWidget {
  const SubscriptionsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final channels = [
      ('GVP Tech Studio', '@gvptech', true),
      ('Cloud Native Foundry', '@cloudnative', true),
      ('Cinema4K HDR Lab', '@cinema4k', false),
      ('SynthWave Live', '@synthwave', false),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('Subscriptions'),
        actions: [
          TextButton(
            onPressed: () => context.push('/channel/@gvptech'),
            child: const Text('All'),
          ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          SizedBox(
            height: 92,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: channels.length,
              separatorBuilder: (_, __) => const SizedBox(width: 14),
              itemBuilder: (context, index) {
                final ch = channels[index];
                return InkWell(
                  onTap: () => context.push('/channel/${ch.$2}'),
                  child: Column(
                    children: [
                      Stack(
                        children: [
                          CircleAvatar(
                            radius: 28,
                            backgroundColor: const Color(0xFF27273A),
                            child: Text(ch.$1.substring(0, 2).toUpperCase()),
                          ),
                          if (ch.$3)
                            Positioned(
                              right: 0,
                              bottom: 0,
                              child: Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 5,
                                  vertical: 1,
                                ),
                                decoration: BoxDecoration(
                                  color: const Color(0xFFE11D48),
                                  borderRadius: BorderRadius.circular(4),
                                ),
                                child: const Text(
                                  'LIVE',
                                  style: TextStyle(
                                    fontSize: 9,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ),
                        ],
                      ),
                      const SizedBox(height: 6),
                      SizedBox(
                        width: 72,
                        child: Text(
                          ch.$1,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          textAlign: TextAlign.center,
                          style: const TextStyle(fontSize: 11),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),
          const Divider(height: 24),
          ListTile(
            contentPadding: EdgeInsets.zero,
            leading: Container(
              width: 110,
              height: 64,
              decoration: BoxDecoration(
                color: const Color(0xFF232334),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(Icons.sensors_rounded, color: Color(0xFFE11D48)),
            ),
            title: const Text('Global Keynote Live Stream - Low Latency LL-HLS'),
            subtitle: const Text('GVP Tech Studio • 14.2K watching now'),
            onTap: () => context.push('/live/live_keynote_01'),
          ),
        ],
      ),
    );
  }
}

class ChannelScreen extends StatelessWidget {
  final String handle;
  const ChannelScreen({super.key, required this.handle});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(handle)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          Row(
            children: [
              const CircleAvatar(
                radius: 36,
                backgroundColor: Color(0xFFE11D48),
                child: Icon(Icons.verified_user_rounded, size: 32),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      handle,
                      style: const TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      '1.48M subscribers • 320 videos',
                      style: TextStyle(color: Colors.white70, fontSize: 13),
                    ),
                  ],
                ),
              ),
              FilledButton(
                onPressed: () {},
                child: const Text('Subscribed'),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key});

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final _controller = TextEditingController();
  String _query = '';

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _controller,
          autofocus: true,
          decoration: const InputDecoration(
            hintText: 'Search videos, creators, live streams...',
            border: InputBorder.none,
          ),
          onChanged: (value) => setState(() => _query = value),
        ),
      ),
      body: ListView(
        children: [
          for (final suggestion in [
            'AV1 hardware decoding benchmark',
            '4K HDR Dolby Vision nature',
            'Flutter 3.47 Media3 ExoPlayer PiP',
            'Live concert 60fps',
          ].where((item) => item.toLowerCase().contains(_query.toLowerCase())))
            ListTile(
              leading: const Icon(Icons.search_rounded),
              title: Text(suggestion),
              onTap: () => context.push(
                '/watch/vid_search_01?title=${Uri.encodeComponent(suggestion)}',
              ),
            ),
        ],
      ),
    );
  }
}

class NotificationsScreen extends StatelessWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Notifications')),
      body: ListView(
        children: const [
          ListTile(
            leading: CircleAvatar(child: Icon(Icons.live_tv_rounded)),
            title: Text('GVP Engineering went live: Edge Transcoding Q&A'),
            subtitle: Text('4 minutes ago'),
          ),
          ListTile(
            leading: CircleAvatar(child: Icon(Icons.offline_pin_rounded)),
            title: Text('Offline DRM license renewed for 3 downloaded videos'),
            subtitle: Text('2 hours ago'),
          ),
        ],
      ),
    );
  }
}

class PremiumScreen extends StatelessWidget {
  const PremiumScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('GVP Premium')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Ad-Free 4K HDR + Encrypted Offline Downloads + Background Play',
              style: TextStyle(fontSize: 22, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 16),
            const ListTile(
              leading: Icon(Icons.download_done_rounded, color: Color(0xFFE11D48)),
              title: Text('Widevine L1 / FairPlay Offline Vault'),
              subtitle: Text('Watch up to 1080p offline on flights and commutes'),
            ),
            const ListTile(
              leading: Icon(Icons.picture_in_picture_alt_rounded, color: Color(0xFFE11D48)),
              title: Text('Native Picture-in-Picture & Background Audio'),
              subtitle: Text('Seamless lock-screen playback with OS media controls'),
            ),
            const Spacer(),
            SizedBox(
              width: double.infinity,
              child: FilledButton(
                onPressed: () => Navigator.of(context).pop(),
                child: const Text('Active: GVP Premium Pro Plan'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class SettingsScreen extends StatelessWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Settings')),
      body: ListView(
        children: [
          SwitchListTile(
            value: true,
            onChanged: (_) {},
            title: const Text('Background & PiP Playback'),
            subtitle: const Text('Continue audio/PiP when switching apps'),
          ),
          SwitchListTile(
            value: true,
            onChanged: (_) {},
            title: const Text('Download over Wi-Fi only'),
            subtitle: const Text('Save mobile data for encrypted VOD downloads'),
          ),
          ListTile(
            title: const Text('Account & OIDC Session'),
            subtitle: const Text('Signed in via ZITADEL PKCE S256'),
            trailing: OutlinedButton(
              onPressed: () => context.push('/login'),
              child: const Text('Switch Account'),
            ),
          ),
        ],
      ),
    );
  }
}

class OnboardingScreen extends StatelessWidget {
  const OnboardingScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(
                Icons.play_circle_fill_rounded,
                size: 80,
                color: Color(0xFFE11D48),
              ),
              const SizedBox(height: 24),
              const Text(
                'Welcome to Global Video Platform',
                textAlign: TextAlign.center,
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 12),
              const Text(
                'Adaptive 4K HDR streaming, instant 60fps Shorts, live broadcasts, and encrypted offline playback.',
                textAlign: TextAlign.center,
                style: TextStyle(color: Colors.white70),
              ),
              const SizedBox(height: 32),
              SizedBox(
                width: double.infinity,
                child: FilledButton(
                  onPressed: () => context.go('/'),
                  child: const Text('Start Watching'),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Sign In')),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.lock_person_rounded, size: 64, color: Color(0xFFE11D48)),
            const SizedBox(height: 16),
            const Text(
              'ZITADEL OIDC 2.0 + PKCE (S256)',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            const Text(
              'Hardware-backed Keychain / Keystore token vault with automatic refresh rotation.',
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.white70),
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              child: FilledButton.icon(
                icon: const Icon(Icons.login_rounded),
                label: const Text('Continue with OIDC Single Sign-On'),
                onPressed: () => context.go('/'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
