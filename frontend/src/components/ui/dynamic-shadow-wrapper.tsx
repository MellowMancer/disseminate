import React, { isValidElement, useEffect, useRef, useState, type ReactNode } from 'react';
import * as NavigationMenuPrimitive from "@radix-ui/react-navigation-menu"

type DynamicShadowWrapperProps = {
    children: ReactNode;
    className?: string;
};

export function DynamicShadowWrapper({ children, className = '' }: DynamicShadowWrapperProps) {
    const ref = useRef<HTMLDivElement>(null);
    const [shadowOffset, setShadowOffset] = useState({ x: 0, y: 0 });
    const [windowWidth, setWindowWidth] = useState(window.innerWidth);


    useEffect(() => {
        function handleResize() {
            setWindowWidth(window.innerWidth);
        }
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    let isNavMenuItem = false;

    if (React.Children.count(children) === 1) {
        const child = React.Children.only(children);
        if (isValidElement(child)) {
            isNavMenuItem = child.type === NavigationMenuPrimitive.Item;
        }
    }

    // Compute scale factor based on Tailwind breakpoints with override for NavMenuItem
    let scaleFactor = windowWidth >= 768 ? 1 : windowWidth >= 640 ? 0.7 : 0.5;
    if (isNavMenuItem) {
        scaleFactor = 0.15;
    }

    useEffect(() => {
        function updateShadowPosition() {
            if (!ref.current) return;

            const rect = ref.current.getBoundingClientRect();

            let lightSourceX = window.innerWidth / 2 - 120;
            let lightSourceY = 0;

            const elementCenterX = rect.left + rect.width / 2;
            const elementCenterY = rect.top + rect.height / 2;

            let offsetX = ((elementCenterX - lightSourceX) / lightSourceX) * 24;
            let offsetY = ((elementCenterY - lightSourceY) / lightSourceY) * 24;
            offsetX = Math.min(Math.max(offsetX, -24), 24)
            offsetY = Math.min(Math.max(offsetY, -24), 24)
            setShadowOffset({ x: offsetX * scaleFactor, y: offsetY * scaleFactor });
        }



        updateShadowPosition();
        window.addEventListener('resize', updateShadowPosition);
        window.addEventListener('scroll', updateShadowPosition, true);

        return () => {
            window.removeEventListener('mousemove', updateShadowPosition);
            window.removeEventListener('resize', updateShadowPosition);
            window.removeEventListener('scroll', updateShadowPosition, true);
        };
    }, [windowWidth, scaleFactor]);

    return (
        <div
            ref={ref}
            className={`${className} transition-shadow duration-300 ease-in-out rounded-md`}
            style={{
                '--shadow': `${shadowOffset.x}px ${shadowOffset.y}px 0px 0px #282055`,
                boxShadow: 'var(--shadow)',
            } as React.CSSProperties}
        >
            {children}
        </div>
    );
}
