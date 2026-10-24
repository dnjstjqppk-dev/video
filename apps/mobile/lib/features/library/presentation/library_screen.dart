import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/storage/offline_download_manager.dart';

class LibraryScreen extends ConsumerStatefulWidget {
  final int initialTabIndex;

  const LibraryScreen({super.key, this.initialTabIndex = 0});

  @override
  ConsumerState<LibraryScreen> createState() => _LibraryScreenState();
}

class _LibraryScreenState extends ConsumerState<LibraryScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(
      length: 4,
      vsync: this,
      initialIndex: widget.initialTabIndex.clamp(0, 3),
    );
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  String _formatMegabytes(int bytes) =>
      '${(bytes / (1024 * 1024)).toStringAsFixed(0)} MB';

  @override
  Widget build(BuildContext context) {
    final offlineState = ref.watch(offlineDownloadManagerProvider);
    final offlineNotifier = ref.read(offlineDownloadManagerProvider.notifier);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Your Library'),
        actions: [
          TextButton.icon(
            icon: const Icon(Icons.workspace_premium_rounded, size: 18),
            label: const Text('Premium'),
            onPressed: () => context.push('/premium'),
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          isScrollable: true,
          tabs: [
            const Tab(text: 'History'),
            const Tab(text: 'Playlists'),
            const Tab(text: 'Watch Later'),
            Tab(
              key: const Key('library_offline_tab'),
              text: 'Offline (${offlineState.records.length})',
            ),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          // 1. Watch History
          ListView(
            padding: const EdgeInsets.all(16),
            children: [
              ListTile(
                leading: const Icon(Icons.history_rounded),
                title: const Text(
                  'Global Video Mesh: Sub-Second LL-HLS & AV1 Edge Transcoding',
                ),
                subtitle: const Text('Watched 82% • Resume from 19:50'),
                onTap: () => context.push('/watch/vid_home_1'),
              ),
            ],
          ),

          // 2. Playlists
          ListView(
            padding: const EdgeInsets.all(16),
            children: const [
              ListTile(
                leading: Icon(Icons.playlist_play_rounded),
                title: Text('Distributed Systems & Video Codecs'),
                subtitle: Text('18 videos • Updated yesterday'),
              ),
              ListTile(
                leading: Icon(Icons.playlist_play_rounded),
                title: Text('4K Dolby Vision Demo Reels'),
                subtitle: Text('9 videos • Public playlist'),
              ),
            ],
          ),

          // 3. Watch Later
          ListView(
            padding: const EdgeInsets.all(16),
            children: [
              ListTile(
                leading: const Icon(Icons.watch_later_outlined),
                title: const Text(
                  'Distributed ClickHouse Analytics: Real-Time QoE Telemetry',
                ),
                subtitle: const Text('Cloud Native Foundry • 31:42'),
                onTap: () => context.push('/watch/vid_home_4'),
              ),
            ],
          ),

          // 4. Premium Encrypted Offline Downloads
          ListView(
            padding: const EdgeInsets.all(16),
            children: [
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(14),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            'Hardware DRM Vault (Widevine L1 / FairPlay)',
                            style: TextStyle(fontWeight: FontWeight.bold),
                          ),
                          Text(
                            '${_formatMegabytes(offlineState.usedBytes)} / ${_formatMegabytes(offlineState.maxQuotaBytes)}',
                            style: const TextStyle(
                              fontSize: 12,
                              color: Colors.white70,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      LinearProgressIndicator(
                        value: (offlineState.usedBytes /
                                offlineState.maxQuotaBytes)
                            .clamp(0.0, 1.0),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 12),
              for (final record in offlineState.records)
                Card(
                  key: Key('offline_item_${record.videoId}'),
                  margin: const EdgeInsets.only(bottom: 10),
                  child: ListTile(
                    leading: const Icon(
                      Icons.enhanced_encryption_rounded,
                      color: Color(0xFFE11D48),
                    ),
                    title: Text(record.title),
                    subtitle: Text(
                      '${record.resolution} • ${_formatMegabytes(record.sizeBytes)} • ${record.status.name.toUpperCase()}',
                    ),
                    trailing: IconButton(
                      icon: const Icon(Icons.delete_outline_rounded),
                      onPressed: () =>
                          offlineNotifier.removeDownload(record.videoId),
                    ),
                    onTap: () async {
                      try {
                        await offlineNotifier
                            .validateAndStartOfflinePlayback(record.videoId);
                        if (!context.mounted) return;
                        context.push(
                          '/watch/${record.videoId}?title=${Uri.encodeComponent(record.title)}&channel=${Uri.encodeComponent(record.channelName)}',
                        );
                      } on OfflineAccessDeniedException catch (e) {
                        if (!context.mounted) return;
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text(e.reason)),
                        );
                      }
                    },
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }
}
