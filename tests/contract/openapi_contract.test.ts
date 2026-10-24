/**
 * OpenAPI 3.1 Contract & Error Envelope Verification Suite
 * Validates `openapi/gvp-v1.yaml` structural compliance and ensures all `/v1` API error responses
 * strictly conform to `{ error: { code, message, request_id, field? } }`.
 */
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as assert from 'node:assert/strict';
import { describe, it } from 'node:test';

export interface StandardApiErrorEnvelope {
  error: {
    code: string;
    message: string;
    request_id: string;
    field?: string;
  };
}

export function validateApiErrorEnvelope(payload: unknown): payload is StandardApiErrorEnvelope {
  if (typeof payload !== 'object' || payload === null || !('error' in payload)) {
    return false;
  }
  const err = (payload as Record<string, unknown>).error;
  if (typeof err !== 'object' || err === null) {
    return false;
  }
  const e = err as Record<string, unknown>;
  if (typeof e.code !== 'string' || e.code.trim().length === 0) return false;
  if (typeof e.message !== 'string' || e.message.trim().length === 0) return false;
  if (typeof e.request_id !== 'string' || e.request_id.trim().length === 0) return false;
  if (e.field !== undefined && typeof e.field !== 'string') return false;
  return true;
}

describe('OpenAPI 3.1 Contract & Standard Error Envelope Verification', () => {
  const openApiPath = path.resolve(process.cwd(), 'openapi/gvp-v1.yaml');
  const openApiContent = fs.readFileSync(openApiPath, 'utf-8');

  it('declares OpenAPI 3.1.0 version and production/local v1 server endpoints', () => {
    assert.match(openApiContent, /^openapi:\s*3\.1\.0/m);
    assert.match(openApiContent, /url:\s*https:\/\/api\.brand\.com\/v1/);
    assert.match(openApiContent, /url:\s*http:\/\/localhost:8080\/v1/);
  });

  it('defines mandatory core `/v1` paths for identity, playback, feeds, uploads, and analytics', () => {
    const requiredPaths = [
      '/me:',
      '/videos/{id}/playback:',
      '/feed/home:',
      '/feed/shorts:',
      '/uploads:',
      '/events/analytics:',
    ];
    for (const route of requiredPaths) {
      assert.ok(
        openApiContent.includes(route),
        `Expected OpenAPI spec to contain path definition ${route}`
      );
    }
  });

  it('enforces BearerAuth JWT security scheme and standard ApiError schema', () => {
    assert.match(openApiContent, /BearerAuth:/);
    assert.match(openApiContent, /bearerFormat:\s*JWT/);
    assert.match(openApiContent, /ApiError:/);
    assert.match(openApiContent, /required:\s*\[code,\s*message,\s*request_id\]/);
  });

  it('validates canonical error envelope structure `{ error: { code, message, request_id, field? } }`', () => {
    const validUnauthorized: StandardApiErrorEnvelope = {
      error: {
        code: 'UNAUTHENTICATED',
        message: 'Authentication required',
        request_id: 'req_01J9X8Z2M4K1',
      },
    };
    const validFieldValidation: StandardApiErrorEnvelope = {
      error: {
        code: 'INVALID_ARGUMENT',
        message: 'Title is required',
        request_id: 'req_01J9X8Z2M4K2',
        field: 'title',
      },
    };

    assert.equal(validateApiErrorEnvelope(validUnauthorized), true);
    assert.equal(validateApiErrorEnvelope(validFieldValidation), true);

    // Reject malformed error envelopes missing request_id or code
    assert.equal(validateApiErrorEnvelope({ message: 'Flat error not allowed' }), false);
    assert.equal(
      validateApiErrorEnvelope({ error: { code: 'FORBIDDEN_RESOURCE', message: 'Missing request_id' } }),
      false
    );
  });
});
