export async function request(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
    ...options,
  });

  if (!response.ok) {
    let message = 'Request failed';
    let code = '';
    try {
      const body = await response.json();
      message = body?.message || body?.error || body?.error?.message || message;
      code = body?.code || '';
      if (body?.details && typeof body.details === 'object') {
        const firstDetail = Object.values(body.details).find((value) => typeof value === 'string');
        if (firstDetail) {
          message = String(firstDetail);
        }
      }
    } catch {
      // Keep fallback values when the response is not JSON.
    }
    const error = new Error(message);
    error.code = code;
    error.status = response.status;
    throw error;
  }

  if (response.status === 204) {
    return null;
  }

  return response.json();
}
