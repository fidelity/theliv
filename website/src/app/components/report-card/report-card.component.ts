/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, OnInit } from '@angular/core';
import { faExclamationTriangle, faLightbulb, faTimes, faThumbsUp, faShareAlt, faCopy, faDownload } from '@fortawesome/free-solid-svg-icons';
import { KubernetesService } from 'src/app/services/kubernetes.service';
import { SocialUtil } from '../../shared/util/social-util';
const yaml = require('js-yaml');

@Component({
  selector: 'app-report-card',
  templateUrl: './report-card.component.html',
  styleUrls: ['./report-card.component.scss']
})
export class ReportCardComponent implements OnInit {

  resourceGroup: any;
  cat: any;
  count: any;
  ns: any;
  cluster: any;
  selectedResource: any;
  gridToggle = true;
  openDetails = false;
  sortBy = 'time';
  faExclamationTriangle = faExclamationTriangle;
  faLightbulb = faLightbulb;
  faTimes = faTimes;
  faThumbsUp = faThumbsUp;
  faShareAlt = faShareAlt;
  faCopy = faCopy;
  faDownload = faDownload;
  visibleIndex = -1;

  constructor(
    private kubeService: KubernetesService,
    public socialUtil: SocialUtil
  ) { }

  ngOnInit(): void {
    this.kubeService.resourceList$.subscribe((list: any) => {
      this.resourceGroup = list;
      this.count = this.resourceGroup.length;
      console.log(this.resourceGroup)
    });
    this.ns = this.kubeService.selectedNs$.getValue();
    this.cluster = this.kubeService.selectedClusters$.getValue();
    // this.sortData();
  }

  // showPopupItem(index: any): void {
  //   this.selectedResource = this.resourceGroup[index];
  //   if (this.visibleIndex === index) {
  //     this.visibleIndex = -1;
  //     this.gridToggle = false;
  //   } else {
  //     this.selectedResource.template = yaml.safeDump(this.selectedResource.template);
  //     this.visibleIndex = index;
  //     this.gridToggle = true;
  //   }
  // }

  // sortData(): any {
  //   return this.resourceGroup.sort((a: any, b: any) => {
  //     return (new Date(b.createdTime) as any) - (new Date(a.createdTime) as any);
  //   });
  // }
}
