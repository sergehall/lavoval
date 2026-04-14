import type { Metadata } from 'next';

interface CreateMetadataOptions {
  title: string;
  description: string;
  path?: string;
  image?: string;
}

export function createMetadata({
  title,
  description,
  path = '',
  image = '/og-image.png',
}: CreateMetadataOptions): Metadata {
  const url = `https://lavoval.com${path}`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      url,
      images: [{ url: image, width: 1200, height: 630 }],
    },
    twitter: {
      title,
      description,
      images: [image],
    },
  };
}
