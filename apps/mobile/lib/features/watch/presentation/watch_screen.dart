import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../app/router.dart';
import '../../../core/platform/native_media_bridge.dart';
import '../../../core/storage/offline_download_manager.dart';

class ChapterMarker {
  final String title;
  final Duration start;

  const ChapterMarker({required this.title, required this.start});
}

class WatchScreen extends ConsumerStatefulWidget {
  final String videoId;
  final String initialTitle;
  final String initialChannel;

  const WatchScreen({
    super.key,
    required this.videoId,
    required this.initialTitle,
    required this.initialChannel,
  });

  @override
  ConsumerState<WatchScreen> createState() => _WatchScreenState();
}

class _WatchScreenState extends ConsumerState<WatchScreen> {
  static const List<String> _qualities = [
    'Auto (ABR)',
    '1080p60 HDR',
    '720p60',
    '480p',
  ];
  static const List<double> _speeds = [0.5, 1.0, 1.25, 1.5, 2.0];
  static const List<String> _captions = ['Off', 'English (CC)', 'Korean', 'Spanish'];
  static const List<String> _audioTracks = [
    'Original (Stereo / Atmos)',
    'English Dub',
    'Spanish Dub',
  ];

  static const List<ChapterMarker> _chapters = [
    ChapterMarker(title: '00:00 Edge Ingest Architecture', start: Duration.zero),
    ChapterMarker(
      title: '04:30 Adaptive Bitrate Ladder (AV1/HEVC)',
      start: Duration(minutes: 4, seconds: 30),
    ),
    ChapterMarker(
      title: '11:15 Hardware DRM & Offline CENC Vault',
      start: Duration(minutes: 11, seconds: 15),
    ),
    ChapterMarker(
      title: '18:40 Global CDN Anycast Routing',
      start: Duration(minutes: 18, seconds: 40),
    ),
  ];

  bool _isPlaying = true;
  bool _isLandscape = false;
  String _selectedQuality = 'Auto (ABR)';
  double _playbackSpeed = 1.0;
  String _selectedCaption = 'English (CC)';
  String _selectedAudioTrack = 'Original (Stereo / Atmos)';
  double _scrubProgress = 0.22;
  int _selectedChapterIndex = 0;
  bool _isLiked = false;
  int _likeCount = 18420;

  @override
  void initState() {
    super.initState();
    Future.microtask(_syncMediaSession);
  }

  Future<void> _syncMediaSession() async {
    await ref.read(nativeMediaBridgeProvider).updateNowPlayingSession(
          videoId: widget.videoId,
          title: widget.initialTitle,
          channelName: widget.initialChannel,
          durationMs: 24 * 60 * 1000,
          positionMs: (_scrubProgress * 24 * 60 * 1000).round(),
          isPlaying: _isPlaying,
        );
  }

  Future<void> _toggleOrientation() async {
    final nextLandscape = !_isLandscape;
    setState(() => _isLandscape = nextLandscape);
    if (nextLandscape) {
      await SystemChrome.setPreferredOrientations([
        DeviceOrientation.landscapeLeft,
        DeviceOrientation.landscapeRight,
      ]);
    } else {
      await SystemChrome.setPreferredOrientations([
        DeviceOrientation.portraitUp,
        DeviceOrientation.portraitDown,
      ]);
    }
  }

  Future<void> _enterPictureInPicture() async {
    final ok =
        await ref.read(nativeMediaBridgeProvider).enterPictureInPicture();
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          ok
              ? 'Entered native Picture-in-Picture mode'
              : 'PiP not available on this device',
        ),
      ),
    );
  }

  Future<void> _openAirPlayOrCast() async {
    await ref.read(nativeMediaBridgeProvider).showCastOrAirPlayPicker(
          videoId: widget.videoId,
          hlsManifestUrl:
              'https://cdn.gvp.video/hls/${widget.videoId}/master.m3u8',
          title: widget.initialTitle,
        );
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Opened native AirPlay / Google Cast route selector'),
      ),
    );
  }

  void _minimizeToMiniPlayer() {
    ref.read(miniPlayerVideoProvider.notifier).state = MiniPlayerState(
      videoId: widget.videoId,
      title: widget.initialTitle,
      channelName: widget.initialChannel,
      isPlaying: _isPlaying,
      progress: _scrubProgress,
    );
    context.go('/');
  }

  Future<void> _downloadForOffline() async {
    try {
      await ref.read(offlineDownloadManagerProvider.notifier).enqueueDownload(
            videoId: widget.videoId,
            title: widget.initialTitle,
            channelName: widget.initialChannel,
            resolution: _selectedQuality.contains('1080p') ? '1080p' : '720p',
            estimatedSizeBytes: 310 * 1024 * 1024,
            manifestUrl:
                'https://cdn.gvp.video/hls/${widget.videoId}/master.m3u8',
            licenseServerUrl: 'https://drm.gvp.video/v1/license',
          );
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text(
            'Encrypted offline VOD saved to hardware DRM vault (.gvpenc)',
          ),
        ),
      );
    } on OfflineAccessDeniedException catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.reason)),
      );
    }
  }

  void _openCommentsSheet() {
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: const Color(0xFF161622),
      isScrollControlled: true,
      builder: (context) {
        return DraggableScrollableSheet(
          expand: false,
          initialChildSize: 0.6,
          maxChildSize: 0.9,
          builder: (context, scrollController) {
            return ListView(
              controller: scrollController,
              padding: const EdgeInsets.all(16),
              children: const [
                Text(
                  'Comments (1,284)',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                ),
                Divider(height: 24),
                ListTile(
                  leading: CircleAvatar(child: Text('AK')),
                  title: Text('@alex_kernel • 1h ago'),
                  subtitle: Text(
                    'Zero frame drops on 1080p60 HDR with hardware Media3 ExoPlayer!',
                  ),
                ),
                ListTile(
                  leading: CircleAvatar(child: Text('SR')),
                  title: Text('@sara_rust • 45m ago'),
                  subtitle: Text(
                    'Love how the chapter markers sync directly with the scrub bar.',
                  ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  void dispose() {
    SystemChrome.setPreferredOrientations([
      DeviceOrientation.portraitUp,
      DeviceOrientation.portraitDown,
    ]);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final currentChapter = _chapters[_selectedChapterIndex];

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          key: const Key('watch_minimize_miniplayer_btn'),
          tooltip: 'Minimize to Mini-Player',
          icon: const Icon(Icons.keyboard_arrow_down_rounded, size: 28),
          onPressed: _minimizeToMiniPlayer,
        ),
        title: Text(
          widget.initialTitle,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: const TextStyle(fontSize: 15),
        ),
        actions: [
          IconButton(
            key: const Key('watch_airplay_cast_btn'),
            tooltip: 'AirPlay / Google Cast',
            icon: const Icon(Icons.cast_connected_rounded),
            onPressed: _openAirPlayOrCast,
          ),
          IconButton(
            key: const Key('watch_pip_btn'),
            tooltip: 'Picture-in-Picture',
            icon: const Icon(Icons.picture_in_picture_alt_rounded),
            onPressed: _enterPictureInPicture,
          ),
          IconButton(
            key: const Key('watch_orientation_btn'),
            tooltip: 'Toggle Fullscreen Orientation',
            icon: const Icon(Icons.screen_rotation_rounded),
            onPressed: _toggleOrientation,
          ),
        ],
      ),
      body: ListView(
        children: [
          // Adaptive HLS Player Canvas
          AspectRatio(
            aspectRatio: 16 / 9,
            child: Container(
              color: Colors.black,
              child: Stack(
                alignment: Alignment.center,
                children: [
                  Container(
                    decoration: const BoxDecoration(
                      gradient: RadialGradient(
                        colors: [Color(0xFF31102F), Color(0xFF07070C)],
                        radius: 0.9,
                      ),
                    ),
                  ),
                  IconButton(
                    iconSize: 64,
                    icon: Icon(
                      _isPlaying
                          ? Icons.pause_circle_filled_rounded
                          : Icons.play_circle_filled_rounded,
                      color: Colors.white,
                    ),
                    onPressed: () {
                      setState(() => _isPlaying = !_isPlaying);
                      _syncMediaSession();
                    },
                  ),
                  Positioned(
                    top: 10,
                    left: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.black54,
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: Text(
                        'HLS • $_selectedQuality • ${_playbackSpeed}x • CC: $_selectedCaption',
                        style: const TextStyle(fontSize: 11),
                      ),
                    ),
                  ),
                  Positioned(
                    left: 12,
                    right: 12,
                    bottom: 6,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          currentChapter.title,
                          style: const TextStyle(
                            fontSize: 11,
                            color: Colors.white70,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Slider(
                          value: _scrubProgress,
                          activeColor: const Color(0xFFE11D48),
                          onChanged: (val) {
                            setState(() {
                              _scrubProgress = val;
                              _selectedChapterIndex =
                                  (val * (_chapters.length - 1)).round();
                            });
                          },
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),

          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.initialTitle,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 6),
                Text(
                  '${widget.initialChannel} • 284K views • Watch Progress Synced',
                  style: TextStyle(color: Colors.grey.shade400, fontSize: 13),
                ),
                const SizedBox(height: 14),

                // Action Row
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: [
                      ActionChip(
                        avatar: Icon(
                          _isLiked
                              ? Icons.thumb_up_alt_rounded
                              : Icons.thumb_up_off_alt_rounded,
                          size: 18,
                        ),
                        label: Text('$_likeCount'),
                        onPressed: () {
                          setState(() {
                            _isLiked = !_isLiked;
                            _likeCount += _isLiked ? 1 : -1;
                          });
                        },
                      ),
                      const SizedBox(width: 8),
                      ActionChip(
                        key: const Key('watch_comments_btn'),
                        avatar: const Icon(Icons.comment_outlined, size: 18),
                        label: const Text('Comments (1.2K)'),
                        onPressed: _openCommentsSheet,
                      ),
                      const SizedBox(width: 8),
                      ActionChip(
                        key: const Key('watch_offline_download_btn'),
                        avatar: const Icon(Icons.download_rounded, size: 18),
                        label: const Text('Offline Vault'),
                        onPressed: _downloadForOffline,
                      ),
                    ],
                  ),
                ),
                const Divider(height: 28),

                // Player Track & Quality Controls
                Wrap(
                  spacing: 10,
                  runSpacing: 10,
                  children: [
                    DropdownButton<String>(
                      key: const Key('watch_quality_dropdown'),
                      value: _selectedQuality,
                      items: _qualities
                          .map(
                            (q) => DropdownMenuItem(
                              value: q,
                              child: Text('Quality: $q'),
                            ),
                          )
                          .toList(),
                      onChanged: (val) {
                        if (val != null) setState(() => _selectedQuality = val);
                      },
                    ),
                    DropdownButton<double>(
                      key: const Key('watch_speed_dropdown'),
                      value: _playbackSpeed,
                      items: _speeds
                          .map(
                            (s) => DropdownMenuItem(
                              value: s,
                              child: Text('Speed: ${s}x'),
                            ),
                          )
                          .toList(),
                      onChanged: (val) {
                        if (val != null) setState(() => _playbackSpeed = val);
                      },
                    ),
                    DropdownButton<String>(
                      value: _selectedCaption,
                      items: _captions
                          .map(
                            (c) => DropdownMenuItem(
                              value: c,
                              child: Text('Captions: $c'),
                            ),
                          )
                          .toList(),
                      onChanged: (val) {
                        if (val != null) setState(() => _selectedCaption = val);
                      },
                    ),
                    DropdownButton<String>(
                      value: _selectedAudioTrack,
                      items: _audioTracks
                          .map(
                            (a) => DropdownMenuItem(
                              value: a,
                              child: Text('Audio: $a'),
                            ),
                          )
                          .toList(),
                      onChanged: (val) {
                        if (val != null) {
                          setState(() => _selectedAudioTrack = val);
                        }
                      },
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                const Text(
                  'Chapters',
                  style: TextStyle(fontWeight: FontWeight.bold, fontSize: 15),
                ),
                const SizedBox(height: 8),
                for (var i = 0; i < _chapters.length; i++)
                  ListTile(
                    dense: true,
                    contentPadding: EdgeInsets.zero,
                    selected: i == _selectedChapterIndex,
                    leading: const Icon(Icons.bookmark_border_rounded),
                    title: Text(_chapters[i].title),
                    onTap: () {
                      setState(() {
                        _selectedChapterIndex = i;
                        _scrubProgress = i / (_chapters.length - 1);
                      });
                    },
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
