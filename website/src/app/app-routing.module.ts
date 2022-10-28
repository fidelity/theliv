/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { KubePlatformComponent } from './kube-platform/kube-platform.component';
import { ReportCardComponent } from './components/report-card/report-card.component';
import { ResourceGroupContentComponent } from './components/resource-group-content/resource-group-content.component';
import { HomePageComponent } from './components/home-page/home-page.component';

const routes: Routes = [
  {
    path: '', redirectTo: 'home', pathMatch: 'full' 
  },
  {
    path: 'home',
    component: HomePageComponent
  },
  {
    path: 'kubernetes',
    component: KubePlatformComponent,
    children: [
      { path: '', redirectTo: 'overview', pathMatch: 'full' },
      {
        path: 'overview',
        component: ReportCardComponent
      },
      {
        path: 'issue/:id',
        component: ResourceGroupContentComponent
      }
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forRoot(routes, {
      useHash: true
  })],
  exports: [RouterModule]
})
export class AppRoutingModule { }
