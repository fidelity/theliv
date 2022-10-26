/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, Inject } from '@angular/core';
import { OnInit } from '@angular/core';
import { WindowToken } from './shared/util/window';
import { KubernetesService } from './services/kubernetes.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
  isDarkModule = false;
  public env = 'dev';
  configInfo: any;

  constructor( @Inject(WindowToken) private window: Window, private kubeService: KubernetesService) { }

  ngOnInit(): void {
    this.kubeService.getConfigInfo().subscribe((res: any) => {
      if (res) {
        this.configInfo = res;
      }
    }, (err: any) => {
      console.log('Get Config Information Error: ', err);
    });
  }

  themeChange(event: any): void {
    this.isDarkModule = event;
  }

  login() {
    this.window.location.reload();
  }

}


