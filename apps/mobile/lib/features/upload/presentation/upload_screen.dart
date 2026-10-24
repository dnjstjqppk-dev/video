import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

enum MultipartUploadPhase {
  idle,
  uploading,
  paused,
  completed,
  failed,
}

class UploadTaskState {
  final String uploadId;
  final String sourceType; // 'Camera 4K60' or 'Gallery Picker'
  final RangeValues trimRangeSeconds;
  final String title;
  final String description;
  final String visibility; // 'Public', 'Unlisted', 'Private'
  final int selectedThumbnailIndex;
  final int totalChunks;
  final int uploadedChunks;
  final MultipartUploadPhase phase;
  final bool backgroundWorkerActive;

  const UploadTaskState({
    this.uploadId = 'r2_mpu_01',
    this.sourceType = 'Gallery Picker',
    this.trimRangeSeconds = const RangeValues(0, 60),
    this.title = '',
    this.description = '',
    this.visibility = 'Public',
    this.selectedThumbnailIndex = 0,
    this.totalChunks = 8,
    this.uploadedChunks = 0,
    this.phase = MultipartUploadPhase.idle,
    this.backgroundWorkerActive = true,
  });

  double get progress =>
      totalChunks > 0 ? (uploadedChunks / totalChunks).clamp(0.0, 1.0) : 0.0;

  UploadTaskState copyWith({
    String? sourceType,
    RangeValues? trimRangeSeconds,
    String? title,
    String? description,
    String? visibility,
    int? selectedThumbnailIndex,
    int? uploadedChunks,
    MultipartUploadPhase? phase,
    bool? backgroundWorkerActive,
  }) {
    return UploadTaskState(
      uploadId: uploadId,
      sourceType: sourceType ?? this.sourceType,
      trimRangeSeconds: trimRangeSeconds ?? this.trimRangeSeconds,
      title: title ?? this.title,
      description: description ?? this.description,
      visibility: visibility ?? this.visibility,
      selectedThumbnailIndex:
          selectedThumbnailIndex ?? this.selectedThumbnailIndex,
      totalChunks: totalChunks,
      uploadedChunks: uploadedChunks ?? this.uploadedChunks,
      phase: phase ?? this.phase,
      backgroundWorkerActive:
          backgroundWorkerActive ?? this.backgroundWorkerActive,
    );
  }
}

final uploadTaskProvider =
    StateNotifierProvider<UploadTaskNotifier, UploadTaskState>((ref) {
  return UploadTaskNotifier();
});

class UploadTaskNotifier extends StateNotifier<UploadTaskState> {
  UploadTaskNotifier() : super(const UploadTaskState());

  void selectSource(String source) => state = state.copyWith(sourceType: source);

  void updateTrimRange(RangeValues range) =>
      state = state.copyWith(trimRangeSeconds: range);

  void updateMetadata({
    String? title,
    String? description,
    String? visibility,
    int? thumbnailIndex,
  }) {
    state = state.copyWith(
      title: title,
      description: description,
      visibility: visibility,
      selectedThumbnailIndex: thumbnailIndex,
    );
  }

  Future<void> startOrResumeMultipartUpload() async {
    state = state.copyWith(
      phase: MultipartUploadPhase.uploading,
      uploadedChunks: state.uploadedChunks == 0 ? 4 : state.totalChunks,
    );
    await Future<void>.delayed(const Duration(milliseconds: 120));
    if (state.phase == MultipartUploadPhase.uploading) {
      state = state.copyWith(
        uploadedChunks: state.totalChunks,
        phase: MultipartUploadPhase.completed,
      );
    }
  }

  void pauseUpload() {
    if (state.phase == MultipartUploadPhase.uploading) {
      state = state.copyWith(phase: MultipartUploadPhase.paused);
    }
  }

  void resetUpload() {
    state = const UploadTaskState();
  }
}

class UploadScreen extends ConsumerStatefulWidget {
  const UploadScreen({super.key});

  @override
  ConsumerState<UploadScreen> createState() => _UploadScreenState();
}

class _UploadScreenState extends ConsumerState<UploadScreen> {
  final _titleController = TextEditingController(
    text: 'My 4K HDR Creator Studio Vlog',
  );
  final _descriptionController = TextEditingController(
    text: 'Shot on mobile and uploaded via chunked R2 multipart presigned URLs.',
  );

  @override
  void dispose() {
    _titleController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final task = ref.watch(uploadTaskProvider);
    final notifier = ref.read(uploadTaskProvider.notifier);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Creator Upload Studio'),
        actions: [
          if (task.phase == MultipartUploadPhase.completed)
            TextButton(
              onPressed: notifier.resetUpload,
              child: const Text('New Upload'),
            ),
        ],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Source Picker Row
          Row(
            children: [
              Expanded(
                child: OutlinedButton.icon(
                  key: const Key('upload_camera_btn'),
                  icon: const Icon(Icons.videocam_rounded),
                  label: const Text('Camera Capture'),
                  onPressed: () => notifier.selectSource('Camera 4K60'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: OutlinedButton.icon(
                  key: const Key('upload_gallery_btn'),
                  icon: const Icon(Icons.photo_library_rounded),
                  label: const Text('Gallery Picker'),
                  onPressed: () => notifier.selectSource('Gallery Picker'),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // Trim Range Card
          Card(
            child: Padding(
              padding: const EdgeInsets.all(14),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Source: ${task.sourceType} • Trim Clip (${task.trimRangeSeconds.start.round()}s - ${task.trimRangeSeconds.end.round()}s)',
                    style: const TextStyle(fontWeight: FontWeight.w600),
                  ),
                  RangeSlider(
                    values: task.trimRangeSeconds,
                    min: 0,
                    max: 120,
                    divisions: 24,
                    labels: RangeLabels(
                      '${task.trimRangeSeconds.start.round()}s',
                      '${task.trimRangeSeconds.end.round()}s',
                    ),
                    onChanged: notifier.updateTrimRange,
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 16),

          // Metadata Form
          TextField(
            key: const Key('upload_title_field'),
            controller: _titleController,
            decoration: const InputDecoration(
              labelText: 'Title',
              border: OutlineInputBorder(),
            ),
            onChanged: (v) => notifier.updateMetadata(title: v),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _descriptionController,
            maxLines: 3,
            decoration: const InputDecoration(
              labelText: 'Description & Hashtags',
              border: OutlineInputBorder(),
            ),
            onChanged: (v) => notifier.updateMetadata(description: v),
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            value: task.visibility,
            decoration: const InputDecoration(
              labelText: 'Visibility',
              border: OutlineInputBorder(),
            ),
            items: const [
              DropdownMenuItem(value: 'Public', child: Text('Public')),
              DropdownMenuItem(value: 'Unlisted', child: Text('Unlisted')),
              DropdownMenuItem(value: 'Private', child: Text('Private')),
            ],
            onChanged: (v) {
              if (v != null) notifier.updateMetadata(visibility: v);
            },
          ),
          const SizedBox(height: 16),

          // Thumbnail Frame Selector
          const Text(
            'Select Auto-Extracted Frame Thumbnail',
            style: TextStyle(fontWeight: FontWeight.w600),
          ),
          const SizedBox(height: 8),
          Row(
            children: List.generate(3, (idx) {
              final selected = task.selectedThumbnailIndex == idx;
              return Expanded(
                child: GestureDetector(
                  onTap: () => notifier.updateMetadata(thumbnailIndex: idx),
                  child: Container(
                    height: 64,
                    margin: const EdgeInsets.only(right: 8),
                    decoration: BoxDecoration(
                      color: const Color(0xFF222232),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: selected
                            ? const Color(0xFFE11D48)
                            : Colors.transparent,
                        width: 2,
                      ),
                    ),
                    child: Center(child: Text('Frame #${idx + 1}')),
                  ),
                ),
              );
            }),
          ),
          const SizedBox(height: 20),

          // R2 Chunked Multipart Upload Progress
          Card(
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Cloudflare R2 Chunked Multipart Upload',
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                      Text(
                        '${task.uploadedChunks}/${task.totalChunks} parts',
                        style: const TextStyle(color: Colors.white70),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  LinearProgressIndicator(value: task.progress),
                  const SizedBox(height: 8),
                  Text(
                    'Status: ${task.phase.name.toUpperCase()} • Background Task Worker: ACTIVE',
                    style: const TextStyle(fontSize: 12, color: Colors.white70),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      FilledButton.icon(
                        key: const Key('upload_start_btn'),
                        icon: const Icon(Icons.cloud_upload_rounded),
                        label: Text(
                          task.phase == MultipartUploadPhase.paused
                              ? 'Resume Upload'
                              : 'Publish & Upload',
                        ),
                        onPressed: notifier.startOrResumeMultipartUpload,
                      ),
                      const SizedBox(width: 12),
                      if (task.phase == MultipartUploadPhase.uploading)
                        OutlinedButton.icon(
                          icon: const Icon(Icons.pause_rounded),
                          label: const Text('Pause'),
                          onPressed: notifier.pauseUpload,
                        ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
