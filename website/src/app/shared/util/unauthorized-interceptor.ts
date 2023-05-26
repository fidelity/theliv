/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Injectable, Inject } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent, HttpResponse, HttpErrorResponse } from '@angular/common/http';
import { Observable } from 'rxjs';
import { tap } from 'rxjs/operators';
import { WindowToken } from './window';
import { MatDialog } from '@angular/material/dialog';
import { ErrorDialog } from '../errors/error-dialog.component';

@Injectable()
export class UnauthorizedInterceptor implements HttpInterceptor {
    dialogdata: any;

    constructor(@Inject(WindowToken) private window: Window, private dialog: MatDialog, ) { }

    intercept(req: HttpRequest<any>, next: HttpHandler): Observable<HttpEvent<any>> {
        return next.handle(req).pipe(
            tap((event) => {
                console.log(event)
                return;
            }, (err) => {
                if (err instanceof HttpErrorResponse) {
                    if (err.error instanceof ErrorEvent) {
                        console.error("Error Event");
                    } else {
                        console.log(`error status : ${err.status} ${err.statusText}`);
                        switch (err.status) {
                        case 401:      //login
                            const redirectUrl = err.headers.get('X-Location') as string | undefined;
                            if (typeof redirectUrl === 'string') {
                                this.window.location.href = redirectUrl;
                            }
                            break;
                        case 403:     //403 Forbidden
                            this.dialogdata={
                                status: 'Access Denied',
                                message: 'It appears that you do not have adequate permissions on the cluster. Ensure that you are first onboarded to the EKS/RKS cluster with proper permissions.',
                            }
                            break;
                        case 404:       //invalid url 404 Not Found
                            this.dialogdata={
                                status: 'Url Not Found',
                                message: 'It seems the url is invalid. Please make sure your url is valid.'
                            }
                            break;
                        case 500:       //500 Internal Server Error
                            this.dialogdata={
                                status: 'Internal Server Error',
                                message: err.error.message
                            }
                            break;
                        case 502:       //502 Bad Gateway
                            this.dialogdata={
                                status: 'Bad Gateway',
                                message: 'We have a bad gateway. Please contact development team for support.'
                            }
                            break;
                        case 503:       //503 Service Unavailable
                            this.dialogdata={
                                status: 'Service Unavailable',
                                message: err.error.message
                            }
                            break;
                        default:
                            if (err && err.status && (err.error.message || err.statusText)){
                                this.dialogdata={
                                    status: `${err.status}`,
                                    message: `${err.error.message || err.statusText}`,
                                }
                            } else {
                                this.dialogdata={
                                    status: 'Error',
                                    message: 'Woops! There is an error ocurred. Please contact development team for support.'
                                }
                            }
                            
                            break;
                        }
                        if ( this.dialogdata!=null ) {
                            if (this.dialog.openDialogs!=null && this.dialog.openDialogs.length > 0) return;
                            this.dialog.open(ErrorDialog, {
                                data: this.dialogdata,
                            }).afterClosed().pipe();
                        }
                    }
                } else {
                    console.error("some thing else happened");
                }
            })
        );
    }
}
