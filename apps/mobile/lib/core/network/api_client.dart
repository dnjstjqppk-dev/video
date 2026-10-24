import 'dart:math';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../auth/oidc_auth_service.dart';

class ApiErrorException implements Exception {
  final String code;
  final String message;
  final int statusCode;
  final String? requestId;
  final Map<String, dynamic>? details;

  const ApiErrorException({
    required this.code,
    required this.message,
    required this.statusCode,
    this.requestId,
    this.details,
  });

  factory ApiErrorException.fromDioException(DioException err) {
    final statusCode = err.response?.statusCode ?? 500;
    final data = err.response?.data;
    final reqId = err.requestOptions.headers['X-Request-ID'] as String?;

    if (data is Map<String, dynamic>) {
      final errorObj = data['error'] is Map<String, dynamic>
          ? data['error'] as Map<String, dynamic>
          : data;
      return ApiErrorException(
        code: errorObj['code'] as String? ?? 'ERR_API_$statusCode',
        message: errorObj['message'] as String? ??
            err.message ??
            'Unexpected Go API error',
        statusCode: statusCode,
        requestId: errorObj['request_id'] as String? ?? reqId,
        details: errorObj['details'] as Map<String, dynamic>?,
      );
    }

    return ApiErrorException(
      code: 'ERR_NETWORK_$statusCode',
      message: err.message ?? 'Network communication failure',
      statusCode: statusCode,
      requestId: reqId,
    );
  }

  @override
  String toString() =>
      'ApiErrorException(code: $code, status: $statusCode, requestId: $requestId, message: $message)';
}

final apiClientProvider = Provider<ApiClient>((ref) {
  final authService = ref.watch(oidcAuthServiceProvider);
  return ApiClient(authService: authService);
});

class ApiClient {
  final Dio dio;
  final OidcAuthService authService;

  ApiClient({
    required this.authService,
    Dio? customDio,
    String baseUrl = const String.fromEnvironment(
      'GVP_API_BASE_URL',
      defaultValue: 'https://api.gvp.video/v1',
    ),
  }) : dio = customDio ??
            Dio(
              BaseOptions(
                baseUrl: baseUrl,
                connectTimeout: const Duration(seconds: 10),
                receiveTimeout: const Duration(seconds: 15),
                sendTimeout: const Duration(seconds: 15),
                headers: {
                  'Accept': 'application/json',
                  'Content-Type': 'application/json',
                  'X-Client-Platform': 'flutter-mobile-3.47',
                },
              ),
            ) {
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          options.headers['X-Request-ID'] ??= _generateRequestId();
          final token = await authService.getValidAccessToken();
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (err, handler) async {
          final statusCode = err.response?.statusCode;
          final alreadyRetried = err.requestOptions.extra['retried_401'] == true;

          if (statusCode == 401 && !alreadyRetried) {
            final stored = await authService.getStoredTokens();
            if (stored != null && stored.refreshToken.isNotEmpty) {
              final refreshed =
                  await authService.refreshAccessToken(stored.refreshToken);
              if (refreshed != null) {
                final retryOptions = err.requestOptions;
                retryOptions.extra['retried_401'] = true;
                retryOptions.headers['Authorization'] =
                    'Bearer ${refreshed.accessToken}';
                try {
                  final response = await dio.fetch(retryOptions);
                  return handler.resolve(response);
                } on DioException catch (retryErr) {
                  return handler.next(retryErr);
                }
              }
            }
          }
          handler.next(err);
        },
      ),
    );
  }

  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
  }) async {
    try {
      return await dio.get<T>(path, queryParameters: queryParameters);
    } on DioException catch (e) {
      throw ApiErrorException.fromDioException(e);
    }
  }

  Future<Response<T>> post<T>(
    String path, {
    Object? data,
    Map<String, dynamic>? queryParameters,
  }) async {
    try {
      return await dio.post<T>(
        path,
        data: data,
        queryParameters: queryParameters,
      );
    } on DioException catch (e) {
      throw ApiErrorException.fromDioException(e);
    }
  }

  static String _generateRequestId() {
    final rand = Random.secure();
    final millis = DateTime.now().millisecondsSinceEpoch.toRadixString(16);
    final suffix = List<String>.generate(
      8,
      (_) => rand.nextInt(16).toRadixString(16),
    ).join();
    return 'req_${millis}_$suffix';
  }
}
