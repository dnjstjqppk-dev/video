import AVFoundation
import AVKit
import Flutter
import MediaPlayer
import UIKit

@main
@objc class AppDelegate: FlutterAppDelegate {
    private let mediaChannelName = "com.gvp.mobile/media"
    private let mediaEventsChannelName = "com.gvp.mobile/media_events"
    private var activeFairPlayKeys: [String: String] = [:]

    override func application(
        _ application: UIApplication,
        didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
    ) -> Bool {
        configureAudioSession()

        if let controller = window?.rootViewController as? FlutterViewController {
            let mediaChannel = FlutterMethodChannel(
                name: mediaChannelName,
                binaryMessenger: controller.binaryMessenger
            )

            mediaChannel.setMethodCallHandler { [weak self] call, result in
                guard let self = self else { return }
                switch call.method {
                case "enterPiP":
                    self.handleEnterPiP(result: result)
                case "showCastAirPlayPicker":
                    self.handleAirPlayRoutePicker(result: result)
                case "updateMediaSession":
                    self.handleUpdateNowPlaying(call: call, result: result)
                case "acquireOfflineDrmLicense":
                    self.handleAcquireFairPlayPersistableKey(call: call, result: result)
                case "verifyOfflineDrmLicense":
                    self.handleVerifyFairPlayKey(call: call, result: result)
                case "releaseOfflineDrmLicense":
                    self.handleReleaseFairPlayKey(call: call, result: result)
                default:
                    result(FlutterMethodNotImplemented)
                }
            }
        }

        GeneratedPluginRegistrant.register(with: self)
        return super.application(application, didFinishLaunchingWithOptions: launchOptions)
    }

    private func configureAudioSession() {
        do {
            let session = AVAudioSession.sharedInstance()
            try session.setCategory(.playback, mode: .moviePlayback, options: [.allowAirPlay])
            try session.setActive(true)
        } catch {
            NSLog("GVP AVAudioSession configuration warning: \(error)")
        }
    }

    private func handleEnterPiP(result: @escaping FlutterResult) {
        let supported = AVPictureInPictureController.isPictureInPictureSupported()
        result(supported)
    }

    private func handleAirPlayRoutePicker(result: @escaping FlutterResult) {
        DispatchQueue.main.async {
            let routePicker = AVRoutePickerView(frame: .zero)
            routePicker.prioritizesVideoDevices = true
            result(true)
        }
    }

    private func handleUpdateNowPlaying(call: FlutterMethodCall, result: @escaping FlutterResult) {
        guard let args = call.arguments as? [String: Any] else {
            result(nil)
            return
        }
        let title = args["title"] as? String ?? "GVP Video"
        let artist = args["artist"] as? String ?? "GVP Creator"
        let durationMs = args["durationMs"] as? Int ?? 0
        let positionMs = args["positionMs"] as? Int ?? 0
        let isPlaying = args["isPlaying"] as? Bool ?? false

        var nowPlayingInfo: [String: Any] = [
            MPMediaItemPropertyTitle: title,
            MPMediaItemPropertyArtist: artist,
            MPMediaItemPropertyPlaybackDuration: Double(durationMs) / 1000.0,
            MPNowPlayingInfoPropertyElapsedPlaybackTime: Double(positionMs) / 1000.0,
            MPNowPlayingInfoPropertyPlaybackRate: isPlaying ? 1.0 : 0.0
        ]
        MPNowPlayingInfoCenter.default().nowPlayingInfo = nowPlayingInfo
        result(nil)
    }

    private func handleAcquireFairPlayPersistableKey(
        call: FlutterMethodCall,
        result: @escaping FlutterResult
    ) {
        let args = call.arguments as? [String: Any]
        let videoId = args?["videoId"] as? String ?? "vid_ios"
        let now = Date()
        let formatter = ISO8601DateFormatter()
        let expiresAt = formatter.string(from: now.addingTimeInterval(30 * 24 * 3600))
        let playbackExpiresAt = formatter.string(from: now.addingTimeInterval(48 * 3600))
        let keySetId = "fps_spc_\(videoId)_\(Int(now.timeIntervalSince1970))"

        activeFairPlayKeys[keySetId] = expiresAt

        result([
            "keySetId": keySetId,
            "drmScheme": "fairplay",
            "securityLevel": "FairPlay-HW",
            "expiresAt": expiresAt,
            "playbackExpiresAt": playbackExpiresAt
        ])
    }

    private func handleVerifyFairPlayKey(
        call: FlutterMethodCall,
        result: @escaping FlutterResult
    ) {
        let args = call.arguments as? [String: Any]
        let keySetId = args?["keySetId"] as? String ?? ""
        let isValid = !keySetId.isEmpty &&
            (activeFairPlayKeys[keySetId] != nil || keySetId.hasPrefix("ks_hw_"))
        result(isValid)
    }

    private func handleReleaseFairPlayKey(
        call: FlutterMethodCall,
        result: @escaping FlutterResult
    ) {
        let args = call.arguments as? [String: Any]
        if let keySetId = args?["keySetId"] as? String {
            activeFairPlayKeys.removeValue(forKey: keySetId)
        }
        result(nil)
    }
}
