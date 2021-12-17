import { EventEmitter, Input, Output } from '@angular/core';
import { faUser } from '@fortawesome/free-solid-svg-icons';
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { KubernetesService } from '../../services/kubernetes.service';
import { environment } from '../../../environments/environment';

@Component({
  selector: 'app-navigation-bar',
  templateUrl: './navigation-bar.component.html',
  styleUrls: ['./navigation-bar.component.scss']
})
export class NavigationBarComponent implements OnInit {
  @Input() isDarkModule = false;

  @Output()
  themeChange: EventEmitter<boolean> = new EventEmitter<boolean>();

  faUser = faUser;
  user: any;

  constructor(private kubeService: KubernetesService, private router: Router) {
  }
  
  ngOnInit(): void {
    if (environment.production) {
      this.kubeService.getUserInfo().subscribe((res: any) => {
        if (res) {
          this.user = res;
        }
      }, (err: any) => {
        console.log('Get User Information Error: ', err);
      });
    }
  }

  themeChanged(): void {
    this.isDarkModule = !this.isDarkModule;
    this.themeChange.emit(this.isDarkModule);
  }
  cleanData(): void {
    this.router.navigate(['kubernetes']).then(() => {
      window.location.reload();
    });
  }
}
