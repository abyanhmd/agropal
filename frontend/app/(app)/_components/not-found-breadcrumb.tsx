'use client';

import {NOT_FOUND_LABEL, useSetBreadcrumb} from "@/hooks/breadcrumb-state";

// The breadcrumb only exists inside the app shell, so the shared 404 content can't set it itself
export default function NotFoundBreadcrumb() {
    useSetBreadcrumb([{label: NOT_FOUND_LABEL}]);
    return null;
}
