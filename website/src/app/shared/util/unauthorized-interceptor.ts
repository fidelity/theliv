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
                        case 403:     //forbidden
                            this.dialogdata={
                                status: '403 Forbidden',
                                message: 'Please make sure you have the right access.',
                            }
                            break;
                        case 404:       //invalid url
                            this.dialogdata={
                                status: '404 Not Found',
                                message: 'Please make sure your url is valid.'
                            }
                            break;
                        case 500:       //server error
                            this.dialogdata={
                                status: '500 Internal Server Error',
                                message: 'Please contact development team for support.'
                            }
                            break;
                        case 502:       //bad gateway
                            this.dialogdata={
                                status: '502 Bad Gateway',
                                message: 'Please contact development team for support.'
                            }
                            break;
                        default:
                            this.dialogdata={
                                status: `${err.status}`,
                                message: `${err.statusText}`,
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
