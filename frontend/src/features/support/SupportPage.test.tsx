import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, expect, test, vi } from 'vitest';
import { SupportPage } from './SupportPage';
import { t } from '../../shared/i18n';

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
  vi.unstubAllGlobals();
});

test('renders support as a feed with a category select', async () => {
  fetchMock.mockImplementation(async (input: RequestInfo | URL) => {
    const url = String(input);
    if (url.endsWith('/support/threads')) {
      return Response.json({
        items: [{ id: 'thread-1', title: 'Нужна помощь с коленом', status: 'open' }]
      });
    }
    if (url.endsWith('/discussions/threads')) {
      return Response.json({
        items: [{ id: 'discussion-1', title: 'Идеи для завтрака', category: 'nutrition' }]
      });
    }
    return Response.json({});
  });
  vi.stubGlobal('fetch', fetchMock);

  const user = userEvent.setup();
  renderSupportPage();

  expect(await screen.findByRole('heading', { name: t('ru', 'support.feedTitle') })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: t('ru', 'support.feed.privateTab') })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: t('ru', 'support.feed.communityTab') })).toBeInTheDocument();

  await user.click(screen.getByRole('button', { name: t('ru', 'support.feed.communityTab') }));

  expect(await screen.findByRole('combobox', { name: t('ru', 'support.category') })).toBeInTheDocument();
  expect(await screen.findByText('Идеи для завтрака')).toBeInTheDocument();
});
