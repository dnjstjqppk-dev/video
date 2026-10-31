import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class ShortVideoItem {
  final String id;
  final String title;
  final String creatorHandle;
  final String musicTrack;
  final int likes;
  final int comments;

  const ShortVideoItem({
    required this.id,
    required this.title,
    required this.creatorHandle,
    required this.musicTrack,
    required this.likes,
    required this.comments,
  });
}

class PooledShortController {
  final int index;
  final String videoId;
  bool isInitialized;
  bool isPlaying;
  bool isDisposed;

  PooledShortController({
    required this.index,
    required this.videoId,
    this.isInitialized = false,
    this.isPlaying = false,
    this.isDisposed = false,
  });

  void initialize() {
    if (!isDisposed) {
      isInitialized = true;
    }
  }

  void play() {
    if (!isDisposed && isInitialized) {
      isPlaying = true;
    }
  }

  void pause() {
    isPlaying = false;
  }

  void dispose() {
    isPlaying = false;
    isInitialized = false;
    isDisposed = true;
  }
}

/// Manages active video controllers for the 60fps vertical Shorts feed.
/// Strictly retains only `{currentIndex - 1, currentIndex, currentIndex + 1}` (max 3 controllers)
/// and immediately disposes off-screen controllers to prevent OOM and thermal throttling.
class ShortsControllerPool {
  final int totalItems;
  final Map<int, PooledShortController> _activeControllers = {};
  final List<int> disposedHistory = [];

  ShortsControllerPool({required this.totalItems});

  Map<int, PooledShortController> get activeControllers =>
      Map.unmodifiable(_activeControllers);

  int get activeCount => _activeControllers.length;

  void onPageChanged(int currentIndex) {
    final allowedIndices = <int>{
      if (currentIndex - 1 >= 0) currentIndex - 1,
      currentIndex,
      if (currentIndex + 1 < totalItems) currentIndex + 1,
    };

    // Dispose any controller outside [N-1, N, N+1]
    final toRemove = _activeControllers.keys
        .where((idx) => !allowedIndices.contains(idx))
        .toList();
    for (final idx in toRemove) {
      final controller = _activeControllers.remove(idx);
      controller?.dispose();
      disposedHistory.add(idx);
    }

    // Ensure current and N+1 (and N-1) are initialized and preloaded
    for (final idx in allowedIndices) {
      final controller = _activeControllers.putIfAbsent(
        idx,
        () => PooledShortController(
          index: idx,
          videoId: 'short_vid_$idx',
        )..initialize(),
      );
      if (idx == currentIndex) {
        controller.play();
      } else {
        controller.pause();
      }
    }
  }

  void disposeAll() {
    for (final entry in _activeControllers.entries) {
      entry.value.dispose();
      disposedHistory.add(entry.key);
    }
    _activeControllers.clear();
  }
}

class ShortsScreen extends ConsumerStatefulWidget {
  const ShortsScreen({super.key});

  @override
  ConsumerState<ShortsScreen> createState() => _ShortsScreenState();
}

class _ShortsScreenState extends ConsumerState<ShortsScreen> {
  static final List<ShortVideoItem> _shorts = List.generate(
    12,
    (i) => ShortVideoItem(
      id: 'short_vid_$i',
      title:
          '60fps Vertical AV1 Hardware Decode Showcase #${i + 1} #Flutter347 #Shorts',
      creatorHandle: '@gvp_creator_${i + 1}',
      musicTrack: 'Original Spatial Audio - GVP Studio #${i + 1}',
      likes: 4200 + (i * 310),
      comments: 190 + (i * 15),
    ),
  );

  late final PageController _pageController;
  late final ShortsControllerPool _pool;
  bool _lowBandwidthMode = false;
  final Set<int> _likedIndices = {};
  int? _heartBurstIndex;

  @override
  void initState() {
    super.initState();
    _pageController = PageController();
    _pool = ShortsControllerPool(totalItems: _shorts.length)
      ..onPageChanged(0);
  }

  @override
  void dispose() {
    _pool.disposeAll();
    _pageController.dispose();
    super.dispose();
  }

  void _triggerDoubleTapLike(int index) {
    setState(() {
      _likedIndices.add(index);
      _heartBurstIndex = index;
    });
    Future<void>.delayed(const Duration(milliseconds: 650), () {
      if (mounted && _heartBurstIndex == index) {
        setState(() => _heartBurstIndex = null);
      }
    });
  }

  void _openQuickComments(ShortVideoItem item) {
    showModalBottomSheet<void>(
      context: context,
      backgroundColor: const Color(0xFF14141F),
      builder: (context) => Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Shorts Comments (${item.comments})',
              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const Divider(height: 20),
            const ListTile(
              leading: CircleAvatar(child: Text('JD')),
              title: Text('@j_dev'),
              subtitle: Text('Instant N+1 preload makes swiping buttery smooth!'),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: PageView.builder(
        key: const Key('shorts_page_view'),
        controller: _pageController,
        scrollDirection: Axis.vertical,
        itemCount: _shorts.length,
        onPageChanged: (index) {
          setState(() {
            _pool.onPageChanged(index);
          });
        },
        itemBuilder: (context, index) {
          final item = _shorts[index];
          final isLiked = _likedIndices.contains(index);
          final activePoolSize = _pool.activeCount;

          return GestureDetector(
            key: Key('short_page_$index'),
            onDoubleTap: () => _triggerDoubleTapLike(index),
            child: Stack(
              fit: StackFit.expand,
              children: [
                // Vertical 9:16 Video Surface
                Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        Color.lerp(
                          const Color(0xFF1E1B4B),
                          const Color(0xFF881337),
                          (index % 4) / 3,
                        )!,
                        const Color(0xFF09090E),
                      ],
                      begin: Alignment.topCenter,
                      end: Alignment.bottomCenter,
                    ),
                  ),
                  child: Center(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          _lowBandwidthMode
                              ? Icons.network_check_rounded
                              : Icons.bolt_rounded,
                          size: 64,
                          color: Colors.white54,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          _lowBandwidthMode
                              ? 'Low-Bandwidth Data Saver (360p ABR)'
                              : '60fps Hardware Decoder • Pool Active: $activePoolSize/3',
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),

                // Double-tap heart animation overlay
                if (_heartBurstIndex == index)
                  const Center(
                    child: Icon(
                      Icons.favorite_rounded,
                      color: Color(0xFFE11D48),
                      size: 112,
                    ),
                  ),

                // Top Bar
                SafeArea(
                  child: Align(
                    alignment: Alignment.topRight,
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: FilterChip(
                        label: Text(
                          _lowBandwidthMode ? 'Data Saver ON' : 'HD 60fps',
                        ),
                        selected: _lowBandwidthMode,
                        onSelected: (val) =>
                            setState(() => _lowBandwidthMode = val),
                      ),
                    ),
                  ),
                ),

                // Right-side engagement rail
                Positioned(
                  right: 14,
                  bottom: 40,
                  child: Column(
                    children: [
                      IconButton(
                        key: Key('short_like_btn_$index'),
                        iconSize: 32,
                        icon: Icon(
                          Icons.favorite_rounded,
                          color: isLiked
                              ? const Color(0xFFE11D48)
                              : Colors.white,
                        ),
                        onPressed: () => _triggerDoubleTapLike(index),
                      ),
                      Text('${item.likes + (isLiked ? 1 : 0)}'),
                      const SizedBox(height: 16),
                      IconButton(
                        key: Key('short_comment_btn_$index'),
                        iconSize: 30,
                        icon: const Icon(
                          Icons.chat_bubble_outline_rounded,
                          color: Colors.white,
                        ),
                        onPressed: () => _openQuickComments(item),
                      ),
                      Text('${item.comments}'),
                      const SizedBox(height: 16),
                      IconButton(
                        iconSize: 30,
                        icon: const Icon(Icons.share_rounded, color: Colors.white),
                        onPressed: () {},
                      ),
                      const Text('Share'),
                    ],
                  ),
                ),

                // Bottom metadata overlay
                Positioned(
                  left: 16,
                  right: 84,
                  bottom: 32,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        item.creatorHandle,
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 16,
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        item.title,
                        maxLines: 2,
                        overflow: TextOverflow.ellipsis,
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          const Icon(Icons.music_note_rounded, size: 15),
                          const SizedBox(width: 4),
                          Expanded(
                            child: Text(
                              item.musicTrack,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: const TextStyle(
                                fontSize: 12,
                                color: Colors.white70,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }
}
