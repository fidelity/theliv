/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, Inject } from '@angular/core';
import { OnInit } from '@angular/core';
import { WindowToken } from './shared/util/window';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: ['./app.component.scss']
})
export class AppComponent implements OnInit {
  isDarkModule = false;
  public env = 'dev';

  constructor( @Inject(WindowToken) private window: Window) { }

  ngOnInit(): void {
    
  }

  themeChange(event: any): void {
    this.isDarkModule = event;
  }

  login() {
    this.window.location.reload();
  }

}


