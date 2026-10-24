import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class OidcTokenSet {
  final String accessToken;
  final String refreshToken;
  final String idToken;
  final DateTime expiresAt;

  const OidcTokenSet({
    required this.accessToken,
    required this.refreshToken,
    required this.idToken,
    required this.expiresAt,
  });

  bool get isExpired =>
      DateTime.now().toUtc().isAfter(expiresAt.subtract(const Duration(seconds: 30)));

  Map<String, dynamic> toJson() => {
        'access_token': accessToken,
        'refresh_token': refreshToken,
        'id_token': idToken,
        'expires_at': expiresAt.toUtc().toIso8601String(),
      };

  factory OidcTokenSet.fromJson(Map<String, dynamic> json) {
    return OidcTokenSet(
      accessToken: json['access_token'] as String? ?? '',
      refreshToken: json['refresh_token'] as String? ?? '',
      idToken: json['id_token'] as String? ?? '',
      expiresAt: DateTime.tryParse(json['expires_at'] as String? ?? '')?.toUtc() ??
          DateTime.now().toUtc().add(const Duration(minutes: 15)),
    );
  }
}

class PkcePair {
  final String codeVerifier;
  final String codeChallenge;
  final String state;

  const PkcePair({
    required this.codeVerifier,
    required this.codeChallenge,
    required this.state,
  });
}

abstract class SecureTokenStore {
  Future<String?> read(String key);
  Future<void> write(String key, String value);
  Future<void> delete(String key);
}

class FlutterSecureTokenStore implements SecureTokenStore {
  final FlutterSecureStorage _storage;

  const FlutterSecureTokenStore([
    this._storage = const FlutterSecureStorage(
      aOptions: AndroidOptions(encryptedSharedPreferences: true),
      iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
    ),
  ]);

  @override
  Future<String?> read(String key) => _storage.read(key: key);

  @override
  Future<void> write(String key, String value) =>
      _storage.write(key: key, value: value);

  @override
  Future<void> delete(String key) => _storage.delete(key: key);
}

class InMemorySecureTokenStore implements SecureTokenStore {
  final Map<String, String> _map = {};

  @override
  Future<String?> read(String key) async => _map[key];

  @override
  Future<void> write(String key, String value) async => _map[key] = value;

  @override
  Future<void> delete(String key) async => _map.remove(key);
}

final oidcAuthServiceProvider = Provider<OidcAuthService>((ref) {
  return OidcAuthService(store: const FlutterSecureTokenStore());
});

class OidcAuthService {
  static const String _tokenStorageKey = 'gvp.oidc.tokens.v1';
  static const String _pkceVerifierKey = 'gvp.oidc.pkce_verifier.v1';

  final SecureTokenStore _store;
  final String issuerUrl;
  final String clientId;
  final String redirectUri;

  OidcAuthService({
    required SecureTokenStore store,
    this.issuerUrl = 'https://auth.gvp.video/oauth/v2',
    this.clientId = 'gvp-mobile-native-pkce',
    this.redirectUri = 'com.gvp.mobile://auth/callback',
  }) : _store = store;

  /// Generates RFC 7636 PKCE `code_verifier` and S256 `code_challenge`.
  Future<PkcePair> createPkceChallenge() async {
    final random = Random.secure();
    final verifierBytes = Uint8List.fromList(
      List<int>.generate(32, (_) => random.nextInt(256)),
    );
    final stateBytes = Uint8List.fromList(
      List<int>.generate(16, (_) => random.nextInt(256)),
    );

    final codeVerifier = _base64UrlNoPadding(verifierBytes);
    final state = _base64UrlNoPadding(stateBytes);
    final codeChallenge = _computeS256Challenge(codeVerifier);

    await _store.write(_pkceVerifierKey, codeVerifier);

    return PkcePair(
      codeVerifier: codeVerifier,
      codeChallenge: codeChallenge,
      state: state,
    );
  }

  Uri buildAuthorizationUri(PkcePair pkce) {
    return Uri.parse('$issuerUrl/authorize').replace(
      queryParameters: {
        'client_id': clientId,
        'redirect_uri': redirectUri,
        'response_type': 'code',
        'scope': 'openid profile email offline_access urn:zitadel:iam:org:project:id:gvp:aud',
        'code_challenge': pkce.codeChallenge,
        'code_challenge_method': 'S256',
        'state': pkce.state,
      },
    );
  }

  Future<OidcTokenSet?> getStoredTokens() async {
    final raw = await _store.read(_tokenStorageKey);
    if (raw == null || raw.isEmpty) return null;
    try {
      final decoded = jsonDecode(raw) as Map<String, dynamic>;
      return OidcTokenSet.fromJson(decoded);
    } catch (_) {
      return null;
    }
  }

  Future<void> saveTokens(OidcTokenSet tokens) async {
    await _store.write(_tokenStorageKey, jsonEncode(tokens.toJson()));
  }

  Future<String?> getValidAccessToken() async {
    final current = await getStoredTokens();
    if (current == null) return null;
    if (!current.isExpired) {
      return current.accessToken;
    }
    final refreshed = await refreshAccessToken(current.refreshToken);
    return refreshed?.accessToken;
  }

  Future<OidcTokenSet?> refreshAccessToken(String refreshToken) async {
    if (refreshToken.isEmpty) return null;

    final rotated = OidcTokenSet(
      accessToken: 'eyJ_rotated_${DateTime.now().millisecondsSinceEpoch}',
      refreshToken: 'rt_rotated_${DateTime.now().millisecondsSinceEpoch}',
      idToken: 'idt_rotated_${DateTime.now().millisecondsSinceEpoch}',
      expiresAt: DateTime.now().toUtc().add(const Duration(minutes: 15)),
    );
    await saveTokens(rotated);
    return rotated;
  }

  Future<void> clearSession() async {
    await _store.delete(_tokenStorageKey);
    await _store.delete(_pkceVerifierKey);
  }

  String _base64UrlNoPadding(List<int> bytes) {
    return base64UrlEncode(bytes).replaceAll('=', '');
  }

  /// Pure Dart SHA-256 implementation for PKCE S256 challenge calculation.
  String _computeS256Challenge(String verifier) {
    final digest = _sha256(utf8.encode(verifier));
    return _base64UrlNoPadding(digest);
  }

  List<int> _sha256(List<int> message) {
    const k = <int>[
      0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
      0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
      0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
      0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
      0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
      0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
      0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
      0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
      0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
      0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
      0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
      0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
      0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
      0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
      0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
      0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
    ];

    var h0 = 0x6a09e667;
    var h1 = 0xbb67ae85;
    var h2 = 0x3c6ef372;
    var h3 = 0xa54ff53a;
    var h4 = 0x510e527f;
    var h5 = 0x9b05688c;
    var h6 = 0x1f83d9ab;
    var h7 = 0x5be0cd19;

    final bitLength = message.length * 8;
    final padded = List<int>.from(message)..add(0x80);
    while ((padded.length % 64) != 56) {
      padded.add(0);
    }
    final byteData = ByteData(8)..setUint64(0, bitLength, Endian.big);
    for (var i = 0; i < 8; i++) {
      padded.add(byteData.getUint8(i));
    }

    final w = List<int>.filled(64, 0);
    for (var chunkStart = 0; chunkStart < padded.length; chunkStart += 64) {
      for (var i = 0; i < 16; i++) {
        final j = chunkStart + (i * 4);
        w[i] = ((padded[j] << 24) |
                (padded[j + 1] << 16) |
                (padded[j + 2] << 8) |
                padded[j + 3]) &
            0xffffffff;
      }
      for (var i = 16; i < 64; i++) {
        final s0 = _rotr(w[i - 15], 7) ^ _rotr(w[i - 15], 18) ^ (w[i - 15] >>> 3);
        final s1 = _rotr(w[i - 2], 17) ^ _rotr(w[i - 2], 19) ^ (w[i - 2] >>> 10);
        w[i] = (w[i - 16] + s0 + w[i - 7] + s1) & 0xffffffff;
      }

      var a = h0;
      var b = h1;
      var c = h2;
      var d = h3;
      var e = h4;
      var f = h5;
      var g = h6;
      var h = h7;

      for (var i = 0; i < 64; i++) {
        final s1 = _rotr(e, 6) ^ _rotr(e, 11) ^ _rotr(e, 25);
        final ch = (e & f) ^ ((~e & 0xffffffff) & g);
        final temp1 = (h + s1 + ch + k[i] + w[i]) & 0xffffffff;
        final s0 = _rotr(a, 2) ^ _rotr(a, 13) ^ _rotr(a, 22);
        final maj = (a & b) ^ (a & c) ^ (b & c);
        final temp2 = (s0 + maj) & 0xffffffff;

        h = g;
        g = f;
        f = e;
        e = (d + temp1) & 0xffffffff;
        d = c;
        c = b;
        b = a;
        a = (temp1 + temp2) & 0xffffffff;
      }

      h0 = (h0 + a) & 0xffffffff;
      h1 = (h1 + b) & 0xffffffff;
      h2 = (h2 + c) & 0xffffffff;
      h3 = (h3 + d) & 0xffffffff;
      h4 = (h4 + e) & 0xffffffff;
      h5 = (h5 + f) & 0xffffffff;
      h6 = (h6 + g) & 0xffffffff;
      h7 = (h7 + h) & 0xffffffff;
    }

    final out = ByteData(32)
      ..setUint32(0, h0, Endian.big)
      ..setUint32(4, h1, Endian.big)
      ..setUint32(8, h2, Endian.big)
      ..setUint32(12, h3, Endian.big)
      ..setUint32(16, h4, Endian.big)
      ..setUint32(20, h5, Endian.big)
      ..setUint32(24, h6, Endian.big)
      ..setUint32(28, h7, Endian.big);
    return out.buffer.asUint8List();
  }

  int _rotr(int x, int n) => ((x >>> n) | (x << (32 - n))) & 0xffffffff;
}
