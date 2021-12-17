import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { KubePlatformComponent } from './kube-platform/kube-platform.component';
import { ReportCardComponent } from './components/report-card/report-card.component';
import { ResourceGroupContentComponent } from './components/resource-group-content/resource-group-content.component';
import { KafkaComponent } from './kafka/kafka.component';

const routes: Routes = [
  {
    path: '', redirectTo: 'kubernetes', pathMatch: 'full' 
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
  },
  {
    path: 'kafka',
    component: KafkaComponent
  },
  {
    path: 'stratum',
    component: KafkaComponent
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
