/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { HttpClient, HttpResponse, HttpHeaders } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { catchError } from 'rxjs/operators';
import { BehaviorSubject, Observable, of, throwError } from 'rxjs';
import { AppConfig } from 'src/config/app.config';
import { environment } from '../../environments/environment';

@Injectable({
  providedIn: 'root'
})
export class KubernetesService {

  kubeEndpoint = AppConfig.KubeApiEndpoint;
  public resourceList$: BehaviorSubject<any> = new BehaviorSubject<any>([]);
  public selectedClusters$: BehaviorSubject<any> = new BehaviorSubject<any>('');
  public selectedNs$: BehaviorSubject<any> = new BehaviorSubject<any>('');

  headers: any;

  constructor(private httpClient: HttpClient) {
    if (!environment.production) {
      this.headers = {
        headers: new HttpHeaders({
          local: 'true'
        })
      };
    } else {
      this.headers = undefined;
    }
  }

  public getClusters(): Observable<any> {
    const url = `${this.kubeEndpoint}/clusters`;
    return this.httpClient.get(url, this.headers).pipe(
      catchError(this.handleError)
    );
  }

  public getAllNamespaces(cluster: string): Observable<any> {
    const url = `${this.kubeEndpoint}/clusters/${cluster}/namespaces`;
    return this.httpClient.get(url, this.headers).pipe(
      catchError(this.handleError)
    );
  }

  public getDetects(cluster: string, namespace: string): Observable<any> {
    const url = `${this.kubeEndpoint}/detector/${cluster}/${namespace}/detect`;
    return this.httpClient.get(url, this.headers).pipe(
      catchError(this.handleError)
    );
  }

  public getUserInfo(): Observable<any> {
    const url = `${this.kubeEndpoint}/userinfo`;
    return this.httpClient.get(url, this.headers).pipe(
      catchError(this.handleError)
    );
  }

  public postUserFeedback(msg: string): Observable<any> {
    const url = `${this.kubeEndpoint}/feedbacks`;
    var body = {
      "message": msg
    }
    return this.httpClient.post(url, body, this.headers).pipe(
      catchError(this.handleError)
    );
  }

  private handleError(response: HttpResponse<any> | any): any {
    return throwError(response || 'Service error');
  }
}
