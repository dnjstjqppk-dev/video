import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class LiveChatMessage {
  final String author;
  final String message;
  final bool isSuperChat;

  const LiveChatMessage({
    required this.author,
    required this.message,
    this.isSuperChat = false,
  });
}

class LiveScreen extends ConsumerStatefulWidget {
  final String streamId;
  final bool initialBroadcasterMode;

  const LiveScreen({
    super.key,
    required this.streamId,
    this.initialBroadcasterMode = false,
  });

  @override
  ConsumerState<LiveScreen> createState() => _LiveScreenState();
}

class _LiveScreenState extends ConsumerState<LiveScreen> {
  late bool _isBroadcasterMode;
  bool _isFrontCamera = true;
  bool _isMicMuted = false;
  bool _isReconnecting = false;
  final TextEditingController _chatController = TextEditingController();

  final List<LiveChatMessage> _messages = [
    const LiveChatMessage(
      author: '@Elena_4K',
      message: 'Sub-2 second LL-HLS latency is crystal clear!',
    ),
    const LiveChatMessage(
      author: '@DevOps_Dan',
      message: 'RTMPS hardware HEVC encoder holding steady at 6.5 Mbps 🔥',
      isSuperChat: true,
    ),
  ];

  @override
  void initState() {
    super.initState();
    _isBroadcasterMode = widget.initialBroadcasterMode;
  }

  @override
  void dispose() {
    _chatController.dispose();
    super.dispose();
  }

  Future<void> _simulateNetworkDropAndAutoReconnect() async {
    setState(() => _isReconnecting = true);
    await Future<void>.delayed(const Duration(milliseconds: 500));
    if (mounted) {
      setState(() => _isReconnecting = false);
    }
  }

  void _sendMessage() {
    final text = _chatController.text.trim();
    if (text.isEmpty) return;
    setState(() {
      _messages.add(LiveChatMessage(author: '@you', message: text));
      _chatController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
              decoration: BoxDecoration(
                color: const Color(0xFFE11D48),
                borderRadius: BorderRadius.circular(4),
              ),
              child: const Text(
                'LIVE',
                style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                _isBroadcasterMode
                    ? 'RTMPS Broadcaster (${widget.streamId})'
                    : 'Live Stream (${widget.streamId})',
                style: const TextStyle(fontSize: 15),
              ),
            ),
          ],
        ),
        actions: [
          TextButton.icon(
            icon: const Icon(Icons.swap_horiz_rounded),
            label: Text(_isBroadcasterMode ? 'Viewer Mode' : 'Go Live Mode'),
            onPressed: () =>
                setState(() => _isBroadcasterMode = !_isBroadcasterMode),
          ),
        ],
      ),
      body: Stack(
        fit: StackFit.expand,
        children: [
          // Live Camera / Stream Canvas
          Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                colors: [Color(0xFF1F1135), Color(0xFF0A0A0F)],
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
              ),
            ),
            child: Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    _isBroadcasterMode
                        ? Icons.sensors_rounded
                        : Icons.live_tv_rounded,
                    size: 72,
                    color: const Color(0xFFE11D48),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    _isReconnecting
                        ? 'Network Drop Detected • Auto-Reconnecting RTMPS...'
                        : _isBroadcasterMode
                            ? 'RTMPS Native Hardware Encoder • ${_isFrontCamera ? "Front" : "Rear"} Camera • 1080p60'
                            : 'LL-HLS Edge Stream • 14,820 Concurrent Viewers',
                    textAlign: TextAlign.center,
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                ],
              ),
            ),
          ),

          // Broadcaster Hardware Controls
          if (_isBroadcasterMode)
            Positioned(
              top: 14,
              left: 14,
              right: 14,
              child: Wrap(
                spacing: 8,
                children: [
                  FilterChip(
                    key: const Key('live_camera_flip_btn'),
                    avatar: const Icon(Icons.cameraswitch_rounded, size: 16),
                    label: Text(_isFrontCamera ? 'Front Cam' : 'Rear Cam'),
                    selected: !_isFrontCamera,
                    onSelected: (_) =>
                        setState(() => _isFrontCamera = !_isFrontCamera),
                  ),
                  FilterChip(
                    key: const Key('live_mic_mute_btn'),
                    avatar: Icon(
                      _isMicMuted ? Icons.mic_off_rounded : Icons.mic_rounded,
                      size: 16,
                    ),
                    label: Text(_isMicMuted ? 'Mic Muted' : 'Mic Live'),
                    selected: _isMicMuted,
                    onSelected: (_) =>
                        setState(() => _isMicMuted = !_isMicMuted),
                  ),
                  ActionChip(
                    avatar: const Icon(Icons.wifi_protected_setup_rounded, size: 16),
                    label: const Text('Test Reconnect'),
                    onPressed: _simulateNetworkDropAndAutoReconnect,
                  ),
                ],
              ),
            ),

          // Real-Time WebSocket Chat Overlay
          Positioned(
            left: 12,
            right: 12,
            bottom: 12,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  height: 170,
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: Colors.black54,
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: ListView.builder(
                    itemCount: _messages.length,
                    itemBuilder: (context, index) {
                      final msg = _messages[index];
                      return Padding(
                        padding: const EdgeInsets.symmetric(vertical: 3),
                        child: RichText(
                          text: TextSpan(
                            children: [
                              TextSpan(
                                text: '${msg.author}: ',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  color: msg.isSuperChat
                                      ? const Color(0xFFFBBF24)
                                      : const Color(0xFFE11D48),
                                ),
                              ),
                              TextSpan(text: msg.message),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _chatController,
                        decoration: InputDecoration(
                          hintText: 'Send live WebSocket chat message...',
                          filled: true,
                          fillColor: const Color(0xFF1A1A26),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 14,
                            vertical: 10,
                          ),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(24),
                            borderSide: BorderSide.none,
                          ),
                        ),
                        onSubmitted: (_) => _sendMessage(),
                      ),
                    ),
                    const SizedBox(width: 8),
                    IconButton.filled(
                      icon: const Icon(Icons.send_rounded),
                      onPressed: _sendMessage,
                    ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
