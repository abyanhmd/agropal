'use client';

import React, {useState} from 'react';
import Image from "next/image";
import {Package} from "lucide-react";

// The product's picture, filling its (relative-positioned) container, or the package icon when it
// has none or the URL doesn't load. The image is any URL the user typed, so it isn't optimized:
// that would need every host listed in next.config's images.remotePatterns.
export default function ProductImage({src, sizes, iconClassName}: {src: string, sizes: string, iconClassName?: string}) {
    const [failed, setFailed] = useState<string | null>(null);

    if (!src || failed === src) return <Package className={iconClassName} />;

    // Decorative: the product's name is always shown next to it
    return <Image src={src} alt="" fill unoptimized sizes={sizes} className="object-cover" onError={() => setFailed(src)} />;
}
