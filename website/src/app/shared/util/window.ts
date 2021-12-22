import { InjectionToken } from '@angular/core';

export const WindowToken = new InjectionToken('Window');
export function windowProvider(): any { return window; }
