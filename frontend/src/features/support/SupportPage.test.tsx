import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { SupportPage } from './SupportPage';

const fetchMock = vi.fn();

function renderSupportPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <SupportPage locale="ru" />
    </QueryClientProvider>
  );
}

afterEach(() => {
  fetchMock.mockReset();
});

test('can reply to support and discussion threads', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = String(input);
    if (url.endsWith('/support/threads') && (!init?.method || init.method === 'GET')) {
      return Response.json({ items: [{ id: 'thread-1', title: 'Нужна помощь', status: 'open' }] });
    }
    if (url.endsWith('/discussions/threads') && (!init?.method || init.method === 'GET')) {
      return Response.json({ items: [{ id: 'discussion-1', title: 'Питание', category: 'nutrition' }] });
    }
    if (url.endsWith('/support/threads/thread-1/messages') && init?.method === 'POST') {
      return Response.json({ thread: { id: 'thread-1' } });
    }
    if (url.endsWith('/discussions/threads/discussion-1/replies') && init?.method === 'POST') {
      return Response.json({ thread: { id: 'discussion-1' } });
    }
    if (init?.method === 'POST') {
      return Response.json({});
    }
    return Response.json({ items: [] });
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderSupportPage();

  expect(await screen.findByText('Нужна помощь')).toBeInTheDocument();
  expect(await screen.findByText('Питание')).toBeInTheDocument();

  const replyFields = await screen.findAllByPlaceholderText('Ответ');
  await user.type(replyFields[0], 'Снизьте нагрузку');
  await user.click(screen.getAllByRole('button', { name: 'Ответить' })[0]);

  await user.type(replyFields[1], 'Попробуйте овсянку');
  await user.click(screen.getAllByRole('button', { name: 'Ответить' })[1]);

  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/support/threads/thread-1/messages',
      expect.objectContaining({ method: 'POST', credentials: 'include' })
    )
  );
  await waitFor(() =>
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/v1/discussions/threads/discussion-1/replies',
      expect.objectContaining({ method: 'POST', credentials: 'include' })
    )
  );
});
